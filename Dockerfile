FROM golang:alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o terminal-server .

FROM docker:dind

COPY --from=builder /app/terminal-server /usr/bin/terminal-server

EXPOSE 27950

ENTRYPOINT ["/usr/bin/terminal-server"]