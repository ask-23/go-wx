# go-wx: A Simple Go Weather Station

[![Go Report Card](https://goreportcard.com/badge/github.com/ask-23/go-wx)](https://goreportcard.com/report/github.com/ask-23/go-wx)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ask-23/go-wx)](https://github.com/ask-23/go-wx)
[![License](https://img.shields.io/github/license/ask-23/go-wx)](https://github.com/ask-23/go-wx/blob/main/LICENSE.md)

A lightweight, high-performance weather monitoring system written in Go. This project was inspired by the amazing [WeeWX](https://github.com/weewx/weewx) project and my experiences configuring, customizing and dockerizing it.

## Features

- Data collection from Ecowitt GW1000 devices using interceptor methodology
- Fast, efficient storage with MariaDB/PostgreSQL
- Colorful and elegant web interface
- Simple configuration via YAML files
- Docker support
- Pre-configured publishing to Weather Underground with extensible service architecture
- Support for both Caddy and Nginx web servers
- Multiple pre-packaged implementation scripts

## Quick Start

### Using Docker

```bash
docker-compose up -d
```

### Manual Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Configure your settings in `config/config.yaml`

3. Run the application:
   ```bash
   go run cmd/go-wx/main.go
   ```

## Configuration

See `config/config.yaml` for all available configuration options.

## Architecture

The go-wx system is designed with modularity in mind:

- Data Collection: Intercepts data from Ecowitt GW1000 devices
- Storage: Efficiently stores weather data in MariaDB/PostgreSQL
- Web Interface: Displays current conditions and historical data
- Publishers: Shares data with external services like Weather Underground

## The Future

The guiding principle of go-wx is *simplicity*. The initial version is configured to use the GW1000 for data collection and either MariaDB or PostgreSQL storage, with optional publishing to Weather Underground. That said, it is determinedly agnostic on products and services and you can configure it to support all the usual alternatives. My motivations for writing were primarily to learn more about Go and I may not develop it much further once the docs are finished.

## License

Licensed under the MIT License. See LICENSE.md for the full text.

Copyright © 2025 Alex Korshak