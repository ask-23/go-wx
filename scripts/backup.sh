#!/bin/bash

# Backup script for go-wx weather station

# Exit on error
set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values
BACKUP_DIR="$HOME/go-wx-backups"
DB_TYPE="mariadb"
DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="go_wx"
DB_USER="go_wx"
CONFIG_DIR="/opt/go-wx/config"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Print banner
echo -e "${GREEN}"
echo "  ____           __        __  __"
echo " / ___| ___      \ \      / /__\ \__"
echo "| |  _ / _ \ _____\ \ /\ / / _ \\\\ \/ /"
echo "| |_| | (_) |_____\ V  V / (_) |>  <"
echo " \____|\___/       \_/\_/ \___//_/\_\\"
echo -e "${NC}"
echo "Go Weather Station Backup Script"
echo ""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
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
    --backup-dir)
      BACKUP_DIR="$2"
      shift 2
      ;;
    --config-dir)
      CONFIG_DIR="$2"
      shift 2
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      exit 1
      ;;
  esac
done

# Create backup directory if it doesn't exist
if [ ! -d "$BACKUP_DIR" ]; then
  echo -e "${YELLOW}Creating backup directory: $BACKUP_DIR${NC}"
  mkdir -p "$BACKUP_DIR"
fi

# Backup filename
BACKUP_FILE="$BACKUP_DIR/go-wx-backup-$TIMESTAMP.tar.gz"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
echo -e "${YELLOW}Creating temporary directory: $TEMP_DIR${NC}"

# Backup database
echo -e "${YELLOW}Backing up database...${NC}"
if [ "$DB_TYPE" == "mariadb" ] || [ "$DB_TYPE" == "mysql" ]; then
  # Prompt for database password
  read -sp "Enter database password for user $DB_USER: " DB_PASS
  echo
  
  # Backup MariaDB/MySQL database
  mysqldump -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" > "$TEMP_DIR/database.sql"
  
  if [ $? -ne 0 ]; then
    echo -e "${RED}Database backup failed!${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
  fi
  
elif [ "$DB_TYPE" == "postgres" ] || [ "$DB_TYPE" == "postgresql" ]; then
  # Prompt for database password
  read -sp "Enter database password for user $DB_USER: " DB_PASS
  echo
  
  # Set PGPASSWORD environment variable
  export PGPASSWORD="$DB_PASS"
  
  # Backup PostgreSQL database
  pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -F c -f "$TEMP_DIR/database.dump"
  
  if [ $? -ne 0 ]; then
    echo -e "${RED}Database backup failed!${NC}"
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

# Backup configuration
echo -e "${YELLOW}Backing up configuration...${NC}"
if [ -d "$CONFIG_DIR" ]; then
  cp -r "$CONFIG_DIR" "$TEMP_DIR/config"
else
  echo -e "${RED}Configuration directory not found: $CONFIG_DIR${NC}"
  echo -e "${YELLOW}Continuing without configuration backup...${NC}"
fi

# Create archive
echo -e "${YELLOW}Creating backup archive...${NC}"
tar -czf "$BACKUP_FILE" -C "$TEMP_DIR" .

# Clean up
echo -e "${YELLOW}Cleaning up temporary files...${NC}"
rm -rf "$TEMP_DIR"

# Print backup information
echo -e "${GREEN}Backup completed successfully!${NC}"
echo "Backup file: $BACKUP_FILE"
echo "Backup size: $(du -h "$BACKUP_FILE" | cut -f1)"
echo ""
echo "To restore this backup, use the restore.sh script:"
echo "./restore.sh --backup-file $BACKUP_FILE" 