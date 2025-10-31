# Build
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/telemetry-service main.go

# Runtime
FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/telemetry-service /telemetry-service
EXPOSE 8083
ENTRYPOINT ["/telemetry-service"]