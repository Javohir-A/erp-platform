FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY genproto ./genproto
COPY services/finance-service ./services/finance-service
WORKDIR /src/services/finance-service
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/finance ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/finance .
EXPOSE 50055
ENTRYPOINT ["/app/finance"]
