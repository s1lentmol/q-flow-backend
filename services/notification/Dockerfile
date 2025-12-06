# syntax=docker/dockerfile:1

FROM golang:1.25 AS builder
WORKDIR /app
ENV GOWORK=off

COPY go.work go.work.sum ./
COPY services/notification/go.mod services/notification/go.sum ./services/notification/
COPY protos/go.mod ./protos/

RUN cd services/notification && go mod download

COPY . .

RUN cd services/notification && CGO_ENABLED=0 GOOS=linux go build -o /bin/notification ./cmd/notification

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=builder /bin/notification /app/notification

EXPOSE 44046

ENTRYPOINT ["/app/notification"]
