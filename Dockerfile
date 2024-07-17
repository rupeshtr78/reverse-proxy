# https://hub.docker.com/_/golang/tags
FROM golang:1.22.2-bullseye AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
RUN mkdir -p /logs /config

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
RUN CGO_ENABLED=1 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -a -o reverseproxy ./cmd/main.go

FROM alpine:3.20.1 AS runtime

WORKDIR /app
COPY --from=builder /app/reverseproxy /app/reverseproxy

COPY config/ /config/
COPY --chown=0:0 . .

RUN apk add --no-cache ca-certificates bash sudo tini

# Install certificates
RUN mkdir -p /usr/local/share/ca-certificates && \
    cp config/keystore/targets/k8s/ca.crt /usr/local/share/ca-certificates/ && \
    update-ca-certificates

# Set environment variables with default values
ENV CONFIG_FILE=/config/config.yaml
ENV LOG_LEVEL=info
ENV LOG_FILE_PATH=/logs/proxy.log
ENV LOG_TO_FILE=false
ENV PORT=8080

# Create a non-root user and group with a writeable home directory
ARG USER_NAME=reverseproxy
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN addgroup -g ${USER_GID} user && \
    adduser -u ${USER_UID} -s /bin/bash -G user --disabled-password --gecos "" ${USER_NAME}

# Create directories before changing ownership
RUN mkdir -p /logs && \
    chown -R $USER_UID:$USER_GID /app /config /logs


# Set the non-root user to ensure the container runs as this user
USER $USER_NAME

EXPOSE ${PORT}
CMD ["/app/reverseproxy"]
