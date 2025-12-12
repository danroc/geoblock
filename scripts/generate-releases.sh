#!/bin/bash

set -e

CHANGELOG="CHANGELOG.md"
DRY_RUN="${DRY_RUN:-true}"

# Get all tags sorted by version
tags=$(git tag -l 'v*' | sort -V)

for tag in $tags; do
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
