#!/bin/bash
# verify-build-artifacts.sh
# Verifies that all expected Thunder build artifacts were created for all platforms
#
# Usage: ./scripts/verify-build-artifacts.sh
#
# Exit codes:
#   0 - All artifacts found
#   1 - One or more artifacts missing

set -e

echo "✅ Verifying all build artifacts were created..."

# Define expected platforms
PLATFORMS=("windows:amd64" "linux:amd64" "linux:arm64" "darwin:amd64" "darwin:arm64")

# Track any missing artifacts
MISSING_COUNT=0

for platform in "${PLATFORMS[@]}"; do
  OS="${platform%%:*}"
  ARCH="${platform#*:}"

  # Determine expected file pattern based on OS
  if [ "$OS" = "windows" ]; then
    EXPECTED_FILE="target/dist/thunder-*-win-*.zip"
  elif [ "$OS" = "darwin" ]; then
    EXPECTED_FILE="target/dist/thunder-*-macos-*.zip"
  else
    EXPECTED_FILE="target/dist/thunder-*-$OS-*.zip"
  fi

  # Check if artifact exists
  if ! ls "$EXPECTED_FILE" 1> /dev/null 2>&1; then
    echo "❌ Backend artifact not found for $OS/$ARCH: $EXPECTED_FILE"
    MISSING_COUNT=$((MISSING_COUNT + 1))
  else
    echo "✅ Found backend artifact for $OS/$ARCH"
  fi
done

if [ $MISSING_COUNT -gt 0 ]; then
  echo ""
  echo "❌ $MISSING_COUNT artifact(s) missing!"
  exit 1
fi

echo ""
echo "✅ All backend artifacts verified successfully!"
exit 0
