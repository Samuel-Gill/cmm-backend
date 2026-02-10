# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/matchmaking-service ./cmd/matchmaking-service

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app

COPY --from=builder /out/matchmaking-service /app/matchmaking-service

ENV SERVER_ADDR=:8080
EXPOSE 8080

ENTRYPOINT ["/app/matchmaking-service"]
