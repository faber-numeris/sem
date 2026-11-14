FROM golang:1.26 AS builder

WORKDIR /app

ENV GO111MODULE=on
ENV GOWORK=off

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o application ./cmd

FROM alpine:3.21 AS runtime

WORKDIR /var/app

COPY --from=builder /app/application /var/app/application
COPY config.yaml /var/app/config.yaml

ENV CONFIG_FILE=/var/app/config.

EXPOSE 8080

CMD ["/var/app/application"]
