#!/bin/bash

# MongoDB Dump Script with Backup Cleanup

# MongoDB Connection Details
MONGO_HOST="localhost"
MONGO_PORT="27017"
MONGO_USERNAME="apartment"
MONGO_PASSWORD="$MONGO_PASSWORD" # Read password from environment variable
AUTH_DB="admin"                  # Authentication database, change if necessary

# Backup Directory
BACKUP_DIR="/root/backup"

# Number of days to retain backups
DAYS_TO_RETAIN=1

# Timestamp for Backup Directory
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Check if the MONGO_PASSWORD environment variable is set
if [ -z "$MONGO_PASSWORD" ]; then
    echo "Error: MongoDB password not set. Set the MONGO_PASSWORD environment variable."
    exit 1
fi

# Remove backups older than specified days
find "$BACKUP_DIR" -maxdepth 1 -type d -mtime +$DAYS_TO_RETAIN -exec rm -rf {} \;

# Create Backup Directory
mkdir -p "$BACKUP_DIR/$TIMESTAMP"

# Run MongoDB Dump
mongodump --host "$MONGO_HOST" --port "$MONGO_PORT" --username "$MONGO_USERNAME" --password "$MONGO_PASSWORD" --authenticationDatabase "$AUTH_DB" --out "$BACKUP_DIR/$TIMESTAMP"

# Check if mongodump was successful
if [ $? -eq 0 ]; then
    echo "Backup successful. Directory: $BACKUP_DIR/$TIMESTAMP"
else
    echo "Backup failed. Check the error message above for details."
fi
