FROM golang:1.25 AS builder
WORKDIR /app

COPY go.work go.work.sum ./
COPY services/queue/go.mod services/queue/go.sum ./services/queue/
COPY protos/go.mod ./protos/

RUN cd services/queue && go mod download

COPY . .

RUN cd services/queue && CGO_ENABLED=0 GOOS=linux go build -o /bin/queue ./cmd/queue

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /bin/queue /app/queue

EXPOSE 44045

ENTRYPOINT ["/app/queue"]
