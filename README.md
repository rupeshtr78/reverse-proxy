# Reverse Proxy Server

This repository contains a reverse proxy server implementation built with Go. The server supports HTTP and HTTPS protocols and allows multiple route configuration through a YAML file. Each route runs on seperate go routine allowing port based routing.It also includes Docker support for containerized deployment.

## Table of Contents

- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [Usage](#usage)
- [Docker](#docker)
  - [Building the Docker Image](#building-the-docker-image)
  - [Running the Docker Container](#running-the-docker-container)
  - [Using Docker Compose](#using-docker-compose)
- [Contributing](#contributing)
- [License](#license)

## Features

- Reverse proxy server with HTTP and HTTPS support.
- Route configuration using a YAML file.
- CORS header support for cross-origin requests.
- TLS configuration for HTTPS routes.
- Error logging and handling.
- Support for multiple routes, each running in a separate goroutine.

## Getting Started

### Prerequisites

To run this project, you'll need:

- Go 1.22 or above
- Docker (optional, for containerized execution)
- Docker Compose (optional, for multi-container orchestration)

### Installation

1. Clone the repository:
    ```sh
    git clone <repository-url>
    cd reverseproxy
    ```

2. Download the Go modules:
    ```sh
    go mod download
    ```

3. Build the project:
    ```sh
    go build -o reverseproxy ./cmd/main.go
    ```

### Configuration

The reverse proxy is configured using a YAML file located at `config/config.yaml`. Below is an example configuration:

```yaml
routes:
  - name: "grafana"
    listenHost: "0.0.0.0"
    listenport: 6442
    protocol: "http"
    certFile: "/path/to/certfile.crt"
    keyFile: "/path/to/keyfile.key"
    pattern: "/"
    target:
      name: "grafana-service"
      protocol: "http"
      host: "192.168.1.100"
      port: 3000
      certfile: "/path/to/target/certfile.crt"
      keyfile: "/path/to/target/keyfile.key"
```

## Usage

To run the reverse proxy server:

1. Ensure the configuration file is properly set up.
2. Execute the server:
    ```sh
    ./reverseproxy
    ```

By default, the server will load configuration from `config/config.yaml`, validate routes, and start the proxy for each route defined. Each route will run in a separate goroutine for concurrent request handling.

## Docker

### Building the Docker Image

1. Build the Docker image:
    ```sh
    docker build -t reverseproxy .
    ```

### Running the Docker Container

1. Run the container:
    ```sh
    docker run -p 8080:8080 reverseproxy
    ```

You can adjust the port and other environment variables as needed. The Dockerfile also supports non-root user setup and signal handling with `tini`.

### Using Docker Compose

For multi-container orchestration, you can use Docker Compose.

1. Ensure you have a `docker-compose.yml` file setup. Example:

    ```yaml
    version: '3'
    services:
      reverseproxy:
        build: .
        ports:
          - "8080:8080"
        environment:
          - CONFIG_FILE=/config/config.yaml
          - LOG_LEVEL=info
        volumes:
          - ./config/config.yaml:/config/config.yaml
    ```

2. Build and run the services:
    ```sh
    docker compose up --build
    ```

This will build the Docker image and start the container as specified in your `docker-compose.yml` file. 

 Adjust the configuration file path, environment variables, and other settings as needed.
 Each route will run in a separate goroutine for concurrent request handling. Add the listen ports to docker `ports` section to expose the services to the host machine.

## Contributing

We welcome contributions! Please open an issue or submit a pull request for any improvements or bug fixes.

### Running Tests

Run the tests using the following command:
```sh
go test ./...
```

Ensure all tests pass before submitting a pull request.

## License

This project is licensed under the MIT License. See the `LICENSE` file for more details.

---

For more information or support, please contact [your contact email].