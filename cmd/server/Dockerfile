FROM golang:1.21-alpine AS build

WORKDIR /go/src/apartment-bot

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go install ./cmd/server/...

FROM alpine:latest

RUN wget -qO /usr/local/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/latest/download/grpc_health_probe-linux-amd64 \
    && chmod +x /usr/local/bin/grpc_health_probe

COPY --from=build /go/bin/server /usr/local/bin/run_app

CMD [ "run_app" ]