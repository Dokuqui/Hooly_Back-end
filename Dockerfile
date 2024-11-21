# Stage 1: Build the Go application
FROM golang:1.23 as builder

WORKDIR /app

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o holly-back

# Stage 2: Create a clean image with only the binary
FROM ubuntu:20.04

WORKDIR /app

COPY --from=builder /app/holly-back /app/holly-back

EXPOSE 8080

CMD ["./holly-back"]