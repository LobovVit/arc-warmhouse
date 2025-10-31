# Build
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/device-service main.go

# Runtime
FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/device-service /device-service
EXPOSE 8082
ENTRYPOINT ["/device-service"]