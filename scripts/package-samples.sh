#!/bin/bash
# package-samples.sh
# Builds and packages sample apps for a specified platform, then verifies the artifacts
#
# Usage: ./scripts/package-samples.sh <OS> <ARCH>
#
# Arguments:
#   OS   - Operating system (linux, win, macos)
#   ARCH - Architecture (x64, arm64)
#
# Examples:
#   ./scripts/package-samples.sh linux x64
#   ./scripts/package-samples.sh macos arm64
#   ./scripts/package-samples.sh win x64
#
# Exit codes:
#   0 - Success: samples built, packaged, and verified
#   1 - Failure: invalid arguments, build failed, or verification failed

set -e

# Check arguments
OS="${1}"
ARCH="${2}"

if [ -z "$OS" ] || [ -z "$ARCH" ]; then
  echo "❌ Error: Missing required arguments"
  echo ""
  echo "Usage: $0 <OS> <ARCH>"
  echo ""
  echo "Arguments:"
  echo "  OS   - Operating system (linux, win, macos)"
  echo "  ARCH - Architecture (x64, arm64)"
  echo ""
  echo "Examples:"
  echo "  $0 linux x64"
  echo "  $0 macos arm64"
  exit 1
fi

echo "📦 Building and packaging samples for $OS/$ARCH..."
echo ""

# Build samples (only once, shared for all platforms)
# Check if samples are already built by looking for the dist directory
if [ ! -d "samples/apps/react-vanilla-sample/dist" ]; then
  echo "🔨 Building sample apps..."
  make build_samples
  echo "✅ Sample apps built successfully"
else
  echo "✅ Sample apps already built, skipping build step"
fi

echo ""

# Package for the specified platform
echo "📦 Packaging samples for $OS/$ARCH..."
make package_samples OS="$OS" ARCH="$ARCH"
echo "✅ Packaging complete"

echo ""

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Verify artifacts were created using the verification script
echo "🔍 Verifying artifacts..."
"$SCRIPT_DIR/verify-sample-artifacts.sh" "$OS" "$ARCH"

echo ""
echo "✅ Sample packaging complete for $OS/$ARCH"
exit 0
