#!/bin/bash

# Raw FotMob API response dumper
# Usage: ./dump-raw-response.sh <match_id>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

if [ "$1" = "--help" ] || [ "$1" = "-h" ] || [ -z "$1" ]; then
    echo "📄 Raw FotMob API Response Dumper"
    echo ""
    echo "Usage: $0 <match_id>"
    echo ""
    echo "This tool dumps the complete raw JSON response from FotMob"
    echo "Use this to inspect the exact API structure and debug parsing"
    echo ""
    echo "Examples:"
    echo "  $0 4813581    Dump raw response for match ID 4813581"
    echo ""
    exit 0
fi

echo "📄 Dumping raw FotMob API response for match ID: $1"
echo ""

go run scripts/dump_raw_response.go "$1"