package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type ServerInfo struct {
	Values map[string]string
}

type ServerStatus struct {
	Values  map[string]string
	Players []PlayerInfo
}

type PlayerInfo struct {
	Score         int
	Ping          int
	Name          string // Name with color codes
	SanitizedName string // Name without color codes
}

type Q3Connector struct {
	host    string
	port    int
	conn    *net.UDPConn
	timeout time.Duration
}

// Create a new Q3Connector instance
func NewQ3Connector(host string, port int) *Q3Connector {
	return &Q3Connector{
		host:    host,
		port:    port,
		timeout: 3 * time.Second,
	}
}

// Establish a UDP connection to the server
func (c *Q3Connector) Connect() error {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return fmt.Errorf("failed to resolve host: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	return nil
}

// Close the UDP connection
func (c *Q3Connector) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Send a command to the server and return the response
func (c *Q3Connector) sendCommand(command string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected to server")
	}

	if err := c.conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return "", fmt.Errorf("failed to set deadline: %w", err)
	}

	// Add the 4-byte header to the command
	commandWithHeader := append([]byte{0xFF, 0xFF, 0xFF, 0xFF}, []byte(command)...)
	_, err := c.conn.Write(commandWithHeader)
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	buffer := make([]byte, 8192)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Skip the 4-byte header in the response
	return string(buffer[4:n]), nil
}

// Get server info
func (c *Q3Connector) GetInfo() (*ServerInfo, error) {
	response, err := c.sendCommand("getinfo\n")
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(response, "infoResponse\n") {
		return nil, fmt.Errorf("invalid getinfo response")
	}

	// Split into lines
	lines := strings.Split(response, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("empty getinfo response")
	}

	return &ServerInfo{
		Values: parseInfoString(lines[1]),
	}, nil
}

// Get server status including players
func (c *Q3Connector) GetStatus() (*ServerStatus, error) {
	response, err := c.sendCommand("getstatus\n")
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(response, "statusResponse\n") {
		return nil, fmt.Errorf("invalid getstatus response")
	}

	// Split into lines
	lines := strings.Split(response, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("empty getstatus response")
	}

	// First line contains server status
	status := &ServerStatus{
		Values:  parseInfoString(lines[1]),
		Players: make([]PlayerInfo, 0),
	}

	// Remaining lines contain player information
	for _, line := range lines[2:] {
		if line == "" {
			continue
		}

		player := parsePlayer(line)
		status.Players = append(status.Players, player)
	}

	return status, nil
}

// Parse a Q3 infostring into a map
func parseInfoString(input string) map[string]string {
	result := make(map[string]string)
	if !strings.HasPrefix(input, "\\") {
		return result
	}

	// Remove the first backslash
	input = input[1:]

	// Split into key-value pairs
	parts := strings.Split(input, "\\")
	for i := 0; i < len(parts)-1; i += 2 {
		key := parts[i]
		value := parts[i+1]
		result[key] = value
	}

	return result
}

// Parse a player info line from status response
func parsePlayer(line string) PlayerInfo {
	player := PlayerInfo{}
	fields := strings.SplitN(line, " ", 3)
	if len(fields) >= 2 {
		fmt.Sscanf(fields[0], "%d", &player.Score)
		fmt.Sscanf(fields[1], "%d", &player.Ping)
		if len(fields) == 3 {
			name := strings.Trim(fields[2], `"`)
			player.Name = name
			player.SanitizedName = sanitizeName(name)
		}
	}
	return player
}

// sanitizeName removes color codes from player names
func sanitizeName(name string) string {
	var result strings.Builder
	for i := 0; i < len(name); i++ {
		if name[i] == '^' && i+1 < len(name) {
			// Skip the caret and the following color code character
			i++
			continue
		}
		result.WriteByte(name[i])
	}
	return result.String()
}
