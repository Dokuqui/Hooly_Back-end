# Stage 1: Build the Go application
FROM golang:1.23 

WORKDIR /app


# Copy the source code
COPY . .

RUN go mod download

# Build the Go binary (output it to /holly-back)
RUN CGO_ENABLED=0 GOOS=linux go build -o holly-back


# Expose the default port (8080)
EXPOSE 8080

# Run the Go binary
CMD ["./holly-back"]