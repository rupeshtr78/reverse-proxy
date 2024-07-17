# https://hub.docker.com/_/golang/tags
FROM golang:1.22.2-bullseye AS builder

ARG TARGETOS
ARG TARGETARCH

RUN mkdir /app
WORKDIR /app


COPY go.mod go.sum ./

RUN set -ex; \
    apt-get update; \
    apt-get install -y --no-install-recommends curl ca-certificates gcc; \
    update-ca-certificates; \
    rm -rf /var/lib/apt/lists/*;

# Download Go modules
RUN go mod download

# Copy the source code
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -a -o proxyserver ./cmd/main.go

FROM alpine:3.20.1 AS runtime

# install bash
RUN apk add --no-cache bash
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN mkdir /app 
WORKDIR /app

RUN mkdir config \
    && mkdir logs


COPY config/ config/

# Install certificates
RUN mkdir -p /usr/local/share/ca-certificates && \
    cp config/keystore/targets/k8s/ca.crt /usr/local/share/ca-certificates/ && \
    update-ca-certificates

# Set environment variables with default values
ENV CONFIG_FILE=/config/config.yaml
ENV LOG_LEVEL=info
ENV LOG_FILE_PATH=/logs/proxy.log
ENV LOG_TO_FILE=false
ENV PORT=6445

# Create a non-root user and group with a writeable home directory
ARG USER_NAME=proxyuser
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN addgroup -g ${USER_GID} user && \
    adduser -u ${USER_UID} -s /bin/bash -G user --disabled-password --gecos "" ${USER_NAME}

# Create directories before changing ownership
RUN chown -R $USER_UID:$USER_GID /app


# Set the non-root user to ensure the container runs as this user
USER $USER_NAME

COPY --chown=$USER_UID:$USER_GID --from=builder /app/proxyserver /app/proxyserver


EXPOSE 1000-65535


CMD ["/app/proxyserver"]
