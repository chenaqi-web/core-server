# Use mirrors that are generally reachable from国内服务器. The image names can
# still be overridden at build time when a different registry is preferred.
ARG GO_IMAGE=docker.m.daocloud.io/library/golang:1.26-alpine
ARG ALPINE_IMAGE=docker.m.daocloud.io/library/alpine:3.21
FROM ${GO_IMAGE} AS builder

WORKDIR /src

# Use a reachable Go module proxy on servers where proxy.golang.org is slow or blocked.
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY .. .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/core-server ./cmd/server

FROM ${ALPINE_IMAGE}

ARG CONFIG_FILE=conf/config.docker.local.yaml

WORKDIR /app

COPY --from=builder /out/core-server ./core-server
COPY ${CONFIG_FILE} ./conf/config.yaml

RUN mkdir -p logs

EXPOSE 8081

CMD ["./core-server"]
