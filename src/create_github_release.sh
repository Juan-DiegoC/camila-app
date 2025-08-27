#!/bin/bash
# GitHub Release Creation Script
# Run this after pushing to GitHub

VERSION="v1.0.0"
REPO="your-username/file-indexer-tui"  # Update with your repo

echo "Creating GitHub release $VERSION..."

# Create release
gh release create $VERSION \
    --title "File Indexer TUI $VERSION" \
    --notes-file RELEASE_NOTES.md \
    release/file-indexer-windows-v1.0.0.zip \
    release/file-indexer-linux-v1.0.0.tar.gz \
    release/file-indexer-macos-v1.0.0.tar.gz \
    release/file-indexer-source-v1.0.0.zip

echo "✅ Release created!"
echo "View at: https://github.com/$REPO/releases/tag/$VERSION"
