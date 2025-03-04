#!/bin/bash

# Installation script for go-wx weather station

# Exit on error
set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Define installation directory
INSTALL_DIR="/opt/go-wx"
CONFIG_DIR="$INSTALL_DIR/config"
LOG_DIR="$INSTALL_DIR/logs"
BIN_DIR="$INSTALL_DIR/bin"
WEB_DIR="$INSTALL_DIR/web"

# Print banner
echo -e "${GREEN}"
echo "  ____           __        __  __"
echo " / ___| ___      \ \      / /__\ \__"
echo "| |  _ / _ \ _____\ \ /\ / / _ \\\\ \/ /"
echo "| |_| | (_) |_____\ V  V / (_) |>  <"
echo " \____|\___/       \_/\_/ \___//_/\_\\"
echo -e "${NC}"
echo "Go Weather Station Installation Script"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Please run as root${NC}"
  exit 1
fi

# Create user and group if they don't exist
echo -e "${YELLOW}Creating go-wx user and group...${NC}"
if ! id -u go-wx > /dev/null 2>&1; then
  useradd -r -s /bin/false go-wx
fi

# Create directories
echo -e "${YELLOW}Creating installation directories...${NC}"
mkdir -p $CONFIG_DIR
mkdir -p $LOG_DIR
mkdir -p $BIN_DIR
mkdir -p $WEB_DIR

# Copy files
echo -e "${YELLOW}Copying files...${NC}"
cp -r ./web/* $WEB_DIR/
cp ./bin/go-wx $BIN_DIR/
cp ./config/config.yaml $CONFIG_DIR/

# Set permissions
echo -e "${YELLOW}Setting permissions...${NC}"
chown -R go-wx:go-wx $INSTALL_DIR
chmod 755 $BIN_DIR/go-wx
chmod 644 $CONFIG_DIR/config.yaml

# Install systemd service
echo -e "${YELLOW}Installing systemd service...${NC}"
cp ./deployments/systemd/go-wx.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable go-wx.service

# Start service
echo -e "${YELLOW}Starting go-wx service...${NC}"
systemctl start go-wx.service

# Check service status
if systemctl is-active --quiet go-wx.service; then
  echo -e "${GREEN}go-wx service is running!${NC}"
else
  echo -e "${RED}Failed to start go-wx service. Check logs with 'journalctl -u go-wx.service'${NC}"
  exit 1
fi

echo -e "${GREEN}Installation complete!${NC}"
echo "Configuration file: $CONFIG_DIR/config.yaml"
echo "Log directory: $LOG_DIR"
echo "Web interface: http://localhost:8080"
echo ""
echo "To check service status: systemctl status go-wx.service"
echo "To view logs: journalctl -u go-wx.service" 