#!/bin/bash
#
# Generate GitHub releases from git tags and CHANGELOG.md
#
# This script extracts release notes from CHANGELOG.md for each git tag and
# creates corresponding GitHub releases using the GitHub CLI (gh).
#
# Requirements:
#   - GitHub CLI (gh) installed and authenticated
#   - CHANGELOG.md in Keep a Changelog format
#   - Git tags in format: vX.Y.Z
#
# Usage:
#   # Dry run (preview what would be created, default)
#   ./scripts/generate-releases.sh
#
#   # Actually create releases
#   DRY_RUN=false ./scripts/generate-releases.sh
#
#   # Use a different changelog file
#   CHANGELOG=HISTORY.md ./scripts/generate-releases.sh
#
# Environment variables:
#   CHANGELOG - Path to changelog file (default: "CHANGELOG.md")
#   DRY_RUN   - Set to "false" to actually create releases (default: "true")

set -euo pipefail

# Check for required dependencies
if ! command -v gh >/dev/null 2>&1; then
    echo "Error: GitHub CLI (gh) is required but not installed or not in PATH." >&2
    echo "Install it from https://cli.github.com/ and try again." >&2
    exit 1
fi

: "${CHANGELOG:=CHANGELOG.md}"
: "${DRY_RUN:=true}"

# Check that the changelog file exists
if [ ! -f "$CHANGELOG" ]; then
    echo "Error: $CHANGELOG not found"
    exit 1
fi

# Print mode indication
if [ "$DRY_RUN" = "true" ]; then
    echo "🔍 Running in DRY RUN mode (no releases will be created)"
    echo "   Set DRY_RUN=false to actually create releases"
    echo ""
else
    echo "✨ Creating releases..."
    echo ""
fi

# Get all tags sorted by version and process each tag safely
git tag -l 'v*' | sort -V | while read -r tag; do
    # Check if release already exists
    if gh release view "$tag" &>/dev/null; then
        echo -e "⏭️  \033[90m$tag already exists, skipping...\033[0m"
        continue
    fi

    # Extract version without 'v' prefix
    version="${tag#v}"

    # Extract changelog section for this version using awk
    # Start at the version header, stop before the next ## header or [links] section
    notes=$(awk "
        /^## \\[$version\\]/ { found=1; next }
        found && /^## \\[/ { exit }
        found && /^\\[Unreleased\\]:/ { exit }
        found { print }
    " "$CHANGELOG" | awk 'NF {p=1} p' | sed '$d')

    if [ -z "$notes" ]; then
        echo "No changelog entry found for $tag, skipping..."
        continue
    fi

    if [ "$DRY_RUN" = "false" ]; then
        echo "✅ Creating release $tag..."
        gh release create "$tag" \
            --title "$tag" \
            --notes "$notes"
    else
        echo ""
        echo -e "📦 \033[1;36m$tag\033[0m"
        echo -e "\033[90m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m"
        echo "$notes"
        echo ""
    fi
done
