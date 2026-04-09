FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY genproto ./genproto
COPY services/hr-service ./services/hr-service
WORKDIR /src/services/hr-service
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/hr ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/hr .
EXPOSE 50052
ENTRYPOINT ["/app/hr"]
