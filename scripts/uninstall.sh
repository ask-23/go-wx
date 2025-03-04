#!/bin/bash

# Uninstallation script for go-wx weather station

# Exit on error
set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Define installation directory
INSTALL_DIR="/opt/go-wx"

# Print banner
echo -e "${RED}"
echo "  ____           __        __  __"
echo " / ___| ___      \ \      / /__\ \__"
echo "| |  _ / _ \ _____\ \ /\ / / _ \\\\ \/ /"
echo "| |_| | (_) |_____\ V  V / (_) |>  <"
echo " \____|\___/       \_/\_/ \___//_/\_\\"
echo -e "${NC}"
echo "Go Weather Station Uninstallation Script"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Please run as root${NC}"
  exit 1
fi

# Confirm uninstallation
echo -e "${YELLOW}This will completely remove go-wx from your system.${NC}"
read -p "Are you sure you want to continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo -e "${GREEN}Uninstallation cancelled.${NC}"
  exit 0
fi

# Stop and disable service
echo -e "${YELLOW}Stopping and disabling go-wx service...${NC}"
if systemctl is-active --quiet go-wx.service; then
  systemctl stop go-wx.service
fi
systemctl disable go-wx.service

# Remove service file
echo -e "${YELLOW}Removing systemd service...${NC}"
rm -f /etc/systemd/system/go-wx.service
systemctl daemon-reload

# Remove installation directory
echo -e "${YELLOW}Removing installation directory...${NC}"
if [ -d "$INSTALL_DIR" ]; then
  rm -rf $INSTALL_DIR
fi

# Ask if user wants to remove the go-wx user
echo -e "${YELLOW}Do you want to remove the go-wx user and group?${NC}"
read -p "Remove user and group? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo -e "${YELLOW}Removing go-wx user and group...${NC}"
  if id -u go-wx > /dev/null 2>&1; then
    userdel go-wx
    echo -e "${GREEN}User and group removed.${NC}"
  else
    echo -e "${YELLOW}User go-wx does not exist.${NC}"
  fi
fi

echo -e "${GREEN}Uninstallation complete!${NC}" 