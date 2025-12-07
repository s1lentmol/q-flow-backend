FROM golang:1.25 AS builder
WORKDIR /app
ENV GOWORK=off
ARG GOPROXY=https://proxy.golang.org,direct
ARG GOSUMDB=sum.golang.org
ENV GOPROXY=$GOPROXY
ENV GOSUMDB=$GOSUMDB

COPY services/auth/go.mod services/auth/go.sum ./services/auth/
COPY protos ./protos

RUN cd services/auth && go mod download

COPY . .

RUN cd services/auth && CGO_ENABLED=0 GOOS=linux go build -o /bin/auth ./cmd/auth

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /bin/auth /app/auth

EXPOSE 44044

ENTRYPOINT ["/app/auth"]
