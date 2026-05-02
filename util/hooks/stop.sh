#!/bin/bash

# Debug hook: dumps the full Stop payload to a log file
# so we can inspect what Claude Code sends us

LOG_FILE="$HOME/.gi-log/debug-stop.log"
mkdir -p "$(dirname "$LOG_FILE")"

echo "--- $(date -u +"%Y-%m-%dT%H:%M:%SZ") ---" >> "$LOG_FILE"
cat >> "$LOG_FILE"
echo "" >> "$LOG_FILE"
