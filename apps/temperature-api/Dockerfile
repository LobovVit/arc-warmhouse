# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod init temperature-api && go mod tidy && go build -o temperature-api main.go

# Runtime
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/temperature-api .
EXPOSE 8081
CMD ["./temperature-api"]