FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY genproto ./genproto
COPY services/procurement-service ./services/procurement-service
WORKDIR /src/services/procurement-service
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/procurement ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/procurement .
EXPOSE 50053
ENTRYPOINT ["/app/procurement"]
