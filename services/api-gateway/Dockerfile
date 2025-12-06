# syntax=docker/dockerfile:1

FROM golang:1.25 AS builder
WORKDIR /app

COPY go.work go.work.sum ./
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY protos/go.mod ./protos/

RUN cd services/api-gateway && go mod download

COPY . .

RUN cd services/api-gateway && CGO_ENABLED=0 GOOS=linux go build -o /bin/api-gateway ./cmd/gateway

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=builder /bin/api-gateway /app/api-gateway

EXPOSE 8080

ENTRYPOINT ["/app/api-gateway"]
