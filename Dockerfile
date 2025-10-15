# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 go build -v -o /go/bin/jka-exporter .

# Final stage
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /go/bin/jka-exporter /

ENTRYPOINT ["/jka-exporter"]
CMD ["-help"]
EXPOSE 8870
