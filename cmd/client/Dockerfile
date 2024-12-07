FROM golang:1.23-alpine AS build

WORKDIR /go/src/apartment-bot

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go install ./cmd/client/...

FROM alpine:latest

COPY --from=build /go/bin/client /usr/local/bin/run_app

CMD [ "run_app" ]