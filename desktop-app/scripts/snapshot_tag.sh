#!/bin/bash
set -e

TAG="${1:-v2.0.0}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"

echo "=========================================================="
echo "⚡ TailRouter - Packaging All-in-One Snapshot for: $TAG"
echo "=========================================================="

TMP_INDEX="/tmp/tailrouter_snapshot_index_$$"
rm -f "$TMP_INDEX"

echo "==> [1/4] Gathering code from all platform branches into folders..."
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=desktop-app/ origin/desktop-app
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=macos/ origin/macos
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=windows/ origin/windows
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=linux/ origin/linux

echo "==> [2/4] Taking root README.md & LICENSE directly from main branch (desktop-app)..."
README_BLOB=$(git hash-object -w README.md)
LICENSE_BLOB=$(git hash-object -w LICENSE)

GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100644 "$README_BLOB" README.md
GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100644 "$LICENSE_BLOB" LICENSE

TREE_ID=$(GIT_INDEX_FILE="$TMP_INDEX" git write-tree)
rm -f "$TMP_INDEX"

echo "==> [3/4] Creating snapshot commit for Tree: $TREE_ID..."
PARENT_COMMIT=$(git rev-parse HEAD)
COMMIT_MSG="release($TAG): all-in-one multi-platform snapshot across desktop-app, macos, windows, linux"
COMMIT_ID=$(echo "$COMMIT_MSG" | git commit-tree "$TREE_ID" -p "$PARENT_COMMIT")

echo "==> [4/4] Updating and pushing tag $TAG to GitHub..."
git tag -fa "$TAG" "$COMMIT_ID" -m "TailRouter $TAG - All-in-One Multi-Platform Snapshot"
git push origin "$TAG" --force

echo "=========================================================="
echo "✅ Finished! Tag $TAG now includes all 4 platform folders:"
echo "   - desktop-app/"
echo "   - macos/"
echo "   - windows/"
echo "   - linux/"
echo "   along with the authoritative English README.md from main."
echo "=========================================================="
