FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY genproto ./genproto
COPY services/auth-service ./services/auth-service
WORKDIR /src/services/auth-service
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/auth ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/auth .
EXPOSE 50051
ENTRYPOINT ["/app/auth"]
