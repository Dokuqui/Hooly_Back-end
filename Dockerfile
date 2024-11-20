FROM golang:1.19

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /holly-back

# Start a new image for the final container
FROM alpine:latest

# Set the working directory inside the new container
WORKDIR /root/

# Copy the Go binary from the builder stage
COPY --from=builder /holly-back .

# we can document in the Dockerfile what ports
# the application is going to listen on by default.
EXPOSE 8080

# Run
CMD ["/holly-back"]