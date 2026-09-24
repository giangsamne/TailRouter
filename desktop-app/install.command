#!/bin/bash
DIR="$(cd "$(dirname "$0")" && pwd)"
bash "$DIR/install.sh"
echo "Press any key to close..."
read -n 1
