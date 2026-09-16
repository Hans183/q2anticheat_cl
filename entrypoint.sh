#!/bin/sh
set -e

# Ensure data and screenshot directories exist
mkdir -p /app/data/screenshots

# Ensure correct ownership for non-root acuser (UID 1000)
chown -R acuser:acuser /app/data

# Execute the application as acuser
exec su-exec acuser "$@"