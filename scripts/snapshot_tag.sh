#!/bin/bash
set -e

TAG="${1:-v2.2.0}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"

echo "=========================================================="
echo "⚡ TailRouter - Packaging All-in-One Snapshot for: $TAG"
echo "=========================================================="

TMP_INDEX="/tmp/tailrouter_snapshot_index_$$"
rm -f "$TMP_INDEX"

echo "==> [1/4] Gathering code from all platform branches into folders..."
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=main/ origin/main
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=macos/ origin/macos
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=windows/ origin/windows
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=linux/ origin/linux

echo "==> [2/4] Adding root README.md, LICENSE, install.sh & install.ps1 pinned to: $TAG..."
README_CONTENT=$(sed -E "s|/v[0-9]+\.[0-9]+\.[0-9]+/|/${TAG}/|g" README.md)
README_BLOB=$(echo "$README_CONTENT" | git hash-object -w --stdin)
LICENSE_BLOB=$(git hash-object -w LICENSE)

INSTALL_SH_CONTENT=$(sed "s|RELEASE_TAG=\".*\"|RELEASE_TAG=\"${TAG}\"|g" install.sh)
INSTALL_SH_BLOB=$(echo "$INSTALL_SH_CONTENT" | git hash-object -w --stdin)

INSTALL_PS1_CONTENT=$(sed "s|\$Tag = \".*\"|\$Tag = \"${TAG}\"|g" install.ps1)
INSTALL_PS1_BLOB=$(echo "$INSTALL_PS1_CONTENT" | git hash-object -w --stdin)

START_SH_BLOB=$(git hash-object -w start.sh)
STOP_SH_BLOB=$(git hash-object -w stop.sh)
INSTALL_SERVICE_SH_BLOB=$(git hash-object -w install-service.sh)

GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100644 "$README_BLOB" README.md
GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100644 "$LICENSE_BLOB" LICENSE
GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100755 "$INSTALL_SH_BLOB" install.sh
GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100644 "$INSTALL_PS1_BLOB" install.ps1
GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100755 "$START_SH_BLOB" start.sh
GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100755 "$STOP_SH_BLOB" stop.sh
GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100755 "$INSTALL_SERVICE_SH_BLOB" install-service.sh

TREE_ID=$(GIT_INDEX_FILE="$TMP_INDEX" git write-tree)
rm -f "$TMP_INDEX"

echo "==> [3/4] Creating snapshot commit for Tree: $TREE_ID..."
PARENT_COMMIT=$(git rev-parse HEAD)
COMMIT_MSG="release($TAG): all-in-one multi-platform snapshot across main, macos, windows, linux"
COMMIT_ID=$(echo "$COMMIT_MSG" | git commit-tree "$TREE_ID" -p "$PARENT_COMMIT")

echo "==> [4/4] Updating and pushing tag $TAG to GitHub..."
git tag -fa "$TAG" "$COMMIT_ID" -m "TailRouter $TAG - All-in-One Multi-Platform Snapshot"
git push origin "$TAG" --force

echo "=========================================================="
echo "✅ Finished! Tag $TAG now includes all 4 platform folders:"
echo "   - main/"
echo "   - macos/"
echo "   - windows/"
echo "   - linux/"
echo "   along with the authoritative English README.md from main."
echo "=========================================================="
