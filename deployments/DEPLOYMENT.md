# Deployment Guide for go-wx Weather Station

You can deploy go-wx in several different ways.

## Assumptions

All deployment scenarios assume you are running a Debian-based Linux OS but are feasible on other platforms. The application has been tested on:

- Debian-based distributions (Debian, Ubuntu)
- RHEL-based distributions (CentOS, Fedora, Amazon Linux)
- Alpine Linux

## Deployment Scenarios

### 1. Standalone Deployment

The simplest deployment method is running the application directly on a server (or your development machine):

```bash
# Build the application
go build -o go-wx cmd/go-wx/main.go

# Create directories for logs and data
mkdir -p /opt/go-wx/logs
mkdir -p /opt/go-wx/data

# Copy the application and configuration
cp go-wx /opt/go-wx/
cp config/config.yaml /opt/go-wx/config.yaml

# Start the application
cd /opt/go-wx
./go-wx -config config.yaml
```

For production, consider using a process manager like supervisor or creating a manual startup script, healthchecks, etc.

### 2. Web Server Integration

Skeleton configs are included for Nginx and Caddy, but any webserver can be configured to work with go-wx.

#### Using NGINX

The NGINX configuration in `deployments/nginx/default.conf` serves static files directly and can be used with the go-wx application running on the same host.

1. Install NGINX: `apt-get install nginx` or `yum install nginx`
2. Copy the configuration: `cp deployments/nginx/default.conf /etc/nginx/conf.d/`
3. Start go-wx on port 8080
4. Reload NGINX: `nginx -s reload`

#### Using Caddy (recommended)

The Caddy configuration in `deployments/caddy/Caddyfile` automatically handles HTTPS certificates.

1. [Install Caddy](https://caddyserver.com/docs/install)
2. Copy the configuration: `cp deployments/caddy/Caddyfile /etc/caddy/Caddyfile`
3. Start go-wx on port 8080
4. Reload Caddy: `systemctl reload caddy`

### 3. Docker Deployment (recommended)

A Docker-based deployment allows for consistent environments across different hosts:

```bash
# Build the Docker image
docker build -t go-wx:latest .

# Run the container
docker run -d \
  --name go-wx \
  -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/data:/app/data \
  go-wx:latest
```

For production, consider using Docker Compose or Kubernetes if you need a new hobby.
Most people would think Kubernetes is a bit much for a Personal Weather Station (PWS) but if you're like that - 
I got you. 

### 4. Database Considerations

SQLite is more than capable of running go-wx but I prefer something a little more robust 
and familiar. And so the app supports different database backends:

- MariaDB/MySQL: Ensure the database server is running and accessible
- PostgreSQL: Configure connection details in the config.yaml file
- SQLite: Ensure the application has write permissions to the database file

Set up the database schema before starting the application:

```bash
# For MariaDB/MySQL
mysql -u username -p weather_db < database/schema.sql

# For PostgreSQL
psql -U username -d weather_db -f database/schema.sql
```

## Monitoring

Consider implementing monitoring using:

- Prometheus for metrics collection
- Grafana for visualization
- AlertManager for alerts

## Backup Strategy

Implement regular backups of:

- Database data
- Configuration files
- Weather data logs

I use [duplicacy](https://github.com/gilbertchen/duplicacy) with a script to 
back up my config and a dump nightly to [storj](https://github.com/storj).

## Security Considerations

- Run the application with limited privileges
- Use proper network segmentation
- Implement HTTPS for all web traffic
- Keep the application and all dependencies updated
