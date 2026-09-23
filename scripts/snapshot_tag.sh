#!/bin/bash
set -e

TAG="${1:-v2.0.0}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"

echo "=========================================================="
echo "⚡ TailRouter - Đóng Gói All-in-One Snapshot Cho Tag: $TAG"
echo "=========================================================="

TMP_INDEX="/tmp/tailrouter_snapshot_index_$$"
rm -f "$TMP_INDEX"

echo "==> [1/4] Đang gom code từ các nhánh vào từng thư mục con..."
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=desktop-app/ origin/desktop-app
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=macos/ origin/macos
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=windows/ origin/windows
GIT_INDEX_FILE="$TMP_INDEX" git read-tree --prefix=linux/ origin/linux

echo "==> [2/4] Tạo README & LICENSE tổng hợp cho Tag..."
TMP_README=$(mktemp)
cat << 'README_EOF' > "$TMP_README"
<div align="center">

# ⚡ TailRouter ($TAG) - All-in-One Multi-Platform Snapshot

**Trọn bộ mã nguồn phát triển của tất cả các hệ điều hành tại phiên bản này**

</div>

---

### 📁 Cấu Trúc Các Thư Mục Trong Phiên Bản Này

Tại mốc Tag **$TAG**, mã nguồn của tất cả các nhánh chuyên biệt được gom lại theo từng thư mục để lưu trữ toàn diện:

| Thư mục | Nhánh tương ứng | Mô tả chi tiết |
| :--- | :--- | :--- |
| **`desktop-app/`** | [\`desktop-app\`](https://github.com/giangsamne/TailRouter/tree/desktop-app) | Nhánh tổng hợp Hub, mã nguồn Go engine chung, web UI nhúng và bộ công cụ build. |
| **`macos/`** | [\`macos\`](https://github.com/giangsamne/TailRouter/tree/macos) | Ứng dụng Menu Bar native viết bằng Swift (\`TailRouter.app\`), Go engine và script cài đặt cho Mac. |
| **`windows/`** | [\`windows\`](https://github.com/giangsamne/TailRouter/tree/windows) | Ứng dụng Khay hệ thống (\`TailRouter.exe\`), Win32 notify icon, Go engine và script cho Windows. |
| **`linux/`** | [\`linux\`](https://github.com/giangsamne/TailRouter/tree/linux) | Bộ cài Desktop launcher, cấu hình dịch vụ \`systemd\` & \`openrc\`, CLI server (x86_64 & ARM64). |

---

> 📜 *Ghi chú: Bản thử nghiệm nguyên bản ban đầu (Python prototype) được lưu trữ riêng tại [**Release v0.1.0**](https://github.com/giangsamne/TailRouter/releases/tag/v0.1.0).*
README_EOF

sed -i "s/\$TAG/$TAG/g" "$TMP_README"

README_BLOB=$(git hash-object -w "$TMP_README")
LICENSE_BLOB=$(git rev-parse origin/desktop-app:LICENSE)
rm -f "$TMP_README"

GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100644 "$README_BLOB" README.md
GIT_INDEX_FILE="$TMP_INDEX" git update-index --add --cacheinfo 100644 "$LICENSE_BLOB" LICENSE

TREE_ID=$(GIT_INDEX_FILE="$TMP_INDEX" git write-tree)
rm -f "$TMP_INDEX"

echo "==> [3/4] Tạo commit Snapshot cho Tree: $TREE_ID..."
PARENT_COMMIT=$(git rev-parse origin/desktop-app)
COMMIT_MSG="release($TAG): all-in-one multi-platform snapshot across desktop-app, macos, windows, linux"
COMMIT_ID=$(echo "$COMMIT_MSG" | git commit-tree "$TREE_ID" -p "$PARENT_COMMIT")

echo "==> [4/4] Cập nhật và đẩy Tag $TAG lên GitHub..."
git tag -fa "$TAG" "$COMMIT_ID" -m "TailRouter $TAG - All-in-One Multi-Platform Snapshot"
git push origin "$TAG" --force

echo "=========================================================="
echo "✅ Hoàn tất! Tag $TAG hiện tại đã gom đầy đủ 4 thư mục:"
echo "   - desktop-app/"
echo "   - macos/"
echo "   - windows/"
echo "   - linux/"
echo "=========================================================="
