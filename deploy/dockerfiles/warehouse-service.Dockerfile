FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY genproto ./genproto
COPY services/warehouse-service ./services/warehouse-service
WORKDIR /src/services/warehouse-service
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/warehouse ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/warehouse .
EXPOSE 50054
ENTRYPOINT ["/app/warehouse"]
