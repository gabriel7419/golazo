#!/bin/bash

# Cache clearing utility for Golazo
# Usage: ./clear-cache.sh [match_id] [--all] [--list] [--team "team name"]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Golazo Cache Clearing Utility"
    echo ""
    echo "Usage:"
    echo "  $0 [match_id]        Clear cache for specific match ID"
    echo "  $0 --all             Clear all cached match details"
    echo "  $0 --list            List all currently cached matches"
    echo "  $0 --team \"Team Name\"  Clear cache for matches containing team name"
    echo "  $0 --force [match_id] Force refresh a match (clear cache + fetch fresh)"
    echo "  $0 --help            Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 12345             Clear cache for match ID 12345"
    echo "  $0 --all             Clear all match details cache"
    echo "  $0 --list            Show all cached matches"
    echo "  $0 --team \"Man City\"   Clear all Man City matches from cache"
    echo "  $0 --force 12345     Force refresh match 12345"
    echo ""
    echo "Notes:"
    echo "  - Cache is in-memory only, cleared when app restarts"
    echo "  - Use --force to test fresh data without waiting for TTL"
    echo ""
    exit 0
fi

if [ "$1" = "--list" ]; then
    echo "📋 Listing currently cached matches..."
    go run scripts/clear_cache.go --list
elif [ "$1" = "--all" ]; then
    echo "Clearing all cached match details..."
    go run scripts/clear_cache.go --all
elif [ "$1" = "--team" ] && [ -n "$2" ]; then
    echo "Clearing cache for matches with team: $2"
    go run scripts/clear_cache.go --team "$2"
elif [ "$1" = "--force" ] && [ -n "$2" ] && [[ "$2" =~ ^[0-9]+$ ]]; then
    echo "🔄 Force refreshing match ID: $2"
    go run scripts/clear_cache.go --force --match "$2"
elif [ -n "$1" ] && [[ "$1" =~ ^[0-9]+$ ]]; then
    echo "Clearing cache for match ID: $1"
    go run scripts/clear_cache.go --match "$1"
else
    echo "❌ Error: Invalid arguments"
    echo ""
    echo "Usage: $0 [match_id] or $0 --all or $0 --list or $0 --team \"team name\" or $0 --force [match_id]"
    echo "Use $0 --help for more information"
    exit 1
fi