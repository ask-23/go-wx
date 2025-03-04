#!/bin/bash

# Restore script for go-wx weather station

# Exit on error
set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values
BACKUP_FILE=""
DB_TYPE="mariadb"
DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="go_wx"
DB_USER="go_wx"
CONFIG_DIR="/opt/go-wx/config"
RESTORE_CONFIG=true

# Print banner
echo -e "${GREEN}"
echo "  ____           __        __  __"
echo " / ___| ___      \ \      / /__\ \__"
echo "| |  _ / _ \ _____\ \ /\ / / _ \\\\ \/ /"
echo "| |_| | (_) |_____\ V  V / (_) |>  <"
echo " \____|\___/       \_/\_/ \___//_/\_\\"
echo -e "${NC}"
echo "Go Weather Station Restore Script"
echo ""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --backup-file)
      BACKUP_FILE="$2"
      shift 2
      ;;
    --db-type)
      DB_TYPE="$2"
      shift 2
      ;;
    --db-host)
      DB_HOST="$2"
      shift 2
      ;;
    --db-port)
      DB_PORT="$2"
      shift 2
      ;;
    --db-name)
      DB_NAME="$2"
      shift 2
      ;;
    --db-user)
      DB_USER="$2"
      shift 2
      ;;
    --config-dir)
      CONFIG_DIR="$2"
      shift 2
      ;;
    --skip-config)
      RESTORE_CONFIG=false
      shift
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      exit 1
      ;;
  esac
done

# Check if backup file is provided
if [ -z "$BACKUP_FILE" ]; then
  echo -e "${RED}Error: Backup file not specified${NC}"
  echo "Usage: $0 --backup-file <path-to-backup-file> [options]"
  exit 1
fi

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
  echo -e "${RED}Error: Backup file not found: $BACKUP_FILE${NC}"
  exit 1
fi

# Create temporary directory
TEMP_DIR=$(mktemp -d)
echo -e "${YELLOW}Creating temporary directory: $TEMP_DIR${NC}"

# Extract backup archive
echo -e "${YELLOW}Extracting backup archive...${NC}"
tar -xzf "$BACKUP_FILE" -C "$TEMP_DIR"

# Check if database backup exists
if [ "$DB_TYPE" == "mariadb" ] || [ "$DB_TYPE" == "mysql" ]; then
  if [ ! -f "$TEMP_DIR/database.sql" ]; then
    echo -e "${RED}Error: Database backup not found in archive${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
  fi
elif [ "$DB_TYPE" == "postgres" ] || [ "$DB_TYPE" == "postgresql" ]; then
  if [ ! -f "$TEMP_DIR/database.dump" ]; then
    echo -e "${RED}Error: Database backup not found in archive${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
  fi
fi

# Restore database
echo -e "${YELLOW}Restoring database...${NC}"
if [ "$DB_TYPE" == "mariadb" ] || [ "$DB_TYPE" == "mysql" ]; then
  # Prompt for database password
  read -sp "Enter database password for user $DB_USER: " DB_PASS
  echo
  
  # Restore MariaDB/MySQL database
  mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$TEMP_DIR/database.sql"
  
  if [ $? -ne 0 ]; then
    echo -e "${RED}Database restore failed!${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
  fi
  
elif [ "$DB_TYPE" == "postgres" ] || [ "$DB_TYPE" == "postgresql" ]; then
  # Prompt for database password
  read -sp "Enter database password for user $DB_USER: " DB_PASS
  echo
  
  # Set PGPASSWORD environment variable
  export PGPASSWORD="$DB_PASS"
  
  # Restore PostgreSQL database
  pg_restore -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$TEMP_DIR/database.dump"
  
  if [ $? -ne 0 ]; then
    echo -e "${RED}Database restore failed!${NC}"
    unset PGPASSWORD
    rm -rf "$TEMP_DIR"
    exit 1
  fi
  
  # Unset PGPASSWORD
  unset PGPASSWORD
else
  echo -e "${RED}Unsupported database type: $DB_TYPE${NC}"
  echo -e "${YELLOW}Supported types: mariadb, mysql, postgres, postgresql${NC}"
  rm -rf "$TEMP_DIR"
  exit 1
fi

# Restore configuration
if [ "$RESTORE_CONFIG" = true ]; then
  echo -e "${YELLOW}Restoring configuration...${NC}"
  if [ -d "$TEMP_DIR/config" ]; then
    # Backup existing config
    if [ -d "$CONFIG_DIR" ]; then
      CONFIG_BACKUP="$CONFIG_DIR.bak.$(date +%s)"
      echo -e "${YELLOW}Backing up existing configuration to $CONFIG_BACKUP${NC}"
      cp -r "$CONFIG_DIR" "$CONFIG_BACKUP"
    fi
    
    # Create config directory if it doesn't exist
    mkdir -p "$CONFIG_DIR"
    
    # Copy config files
    cp -r "$TEMP_DIR/config/"* "$CONFIG_DIR/"
    
    echo -e "${GREEN}Configuration restored successfully!${NC}"
  else
    echo -e "${RED}Configuration not found in backup archive${NC}"
    echo -e "${YELLOW}Continuing without configuration restore...${NC}"
  fi
else
  echo -e "${YELLOW}Skipping configuration restore as requested${NC}"
fi

# Clean up
echo -e "${YELLOW}Cleaning up temporary files...${NC}"
rm -rf "$TEMP_DIR"

# Restart service if it exists
if systemctl is-active --quiet go-wx.service; then
  echo -e "${YELLOW}Restarting go-wx service...${NC}"
  systemctl restart go-wx.service
  
  if systemctl is-active --quiet go-wx.service; then
    echo -e "${GREEN}Service restarted successfully!${NC}"
  else
    echo -e "${RED}Failed to restart service. Please check logs with 'journalctl -u go-wx.service'${NC}"
  fi
fi

echo -e "${GREEN}Restore completed successfully!${NC}" 