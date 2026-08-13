FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY .. .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/core-server ./cmd/server

FROM alpine:3.22

ARG CONFIG_FILE=conf/config.docker.local.yaml

WORKDIR /app

COPY --from=builder /out/core-server ./core-server
COPY ${CONFIG_FILE} ./conf/config.yaml

RUN mkdir -p logs

EXPOSE 8081

CMD ["./core-server"]
