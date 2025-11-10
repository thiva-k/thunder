#!/bin/bash
# ----------------------------------------------------------------------------
# Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
#
# WSO2 LLC. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.
# ----------------------------------------------------------------------------

# Default settings
BACKEND_PORT=${BACKEND_PORT:-8090}
DEBUG_PORT=${DEBUG_PORT:-2345}
DEBUG_MODE=${DEBUG_MODE:-false}
SETUP_MODE=${SETUP_MODE:-false}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --debug)
            DEBUG_MODE=true
            shift
            ;;
        --debug-port)
            DEBUG_PORT="$2"
            shift 2
            ;;
        --port)
            BACKEND_PORT="$2"
            shift 2
            ;;
        --setup)
            SETUP_MODE=true
            shift
            ;;
        --help)
            echo "Thunder Server Startup Script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --debug              Enable debug mode with remote debugging"
            echo "  --port PORT          Set application port (default: 8090)"
            echo "  --debug-port PORT    Set debug port (default: 2345)"
            echo "  --setup              Run initial data setup (automatically disables security temporarily)"
            echo "  --help               Show this help message"
            echo ""
            echo "Setup Mode:"
            echo "  When --setup is used, the server will:"
            echo "  1. Start with security disabled"
            echo "  2. Run the initial data setup script"
            echo "  3. Keep running with security disabled"
            echo "  4. You must restart manually to enable security"
            echo ""
            echo "Examples:"
            echo "  $0                   Start server normally"
            echo "  $0 --setup           Start server and run initial setup"
            echo "  $0 --debug --setup   Start in debug mode and run initial setup"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

set -e  # Exit immediately if a command exits with a non-zero status

# Kill known ports
function kill_port() {
    local port=$1
    lsof -ti tcp:$port | xargs kill -9 2>/dev/null || true
}

# Kill ports before binding
kill_port $BACKEND_PORT
if [ "$DEBUG_MODE" = "true" ]; then
    kill_port $DEBUG_PORT
fi
sleep 1

# Check if Delve is available for debug mode
if [ "$DEBUG_MODE" = "true" ]; then
    # Check for dlv in PATH
    if ! command -v dlv &> /dev/null; then
        echo "❌ Debug mode requires Delve debugger"
        echo ""
        echo "💡 Install Delve using:"
        echo "   go install github.com/go-delve/delve/cmd/dlv@latest"
        echo ""
        echo "🔧 Add Delve to PATH"
        echo ""
        echo "🔧 After installation, run: $0 --debug"
        exit 1
    fi
fi

# Run thunder
if [ "$DEBUG_MODE" = "true" ]; then
    echo "⚡ Starting Thunder Server in DEBUG mode..."
    echo "📝 Application will run on: https://localhost:$BACKEND_PORT"
    echo "🐛 Remote debugger will listen on: localhost:$DEBUG_PORT"
    echo ""
    echo "💡 Connect using remote debugging configuration:"
    echo "   Host: 127.0.0.1, Port: $DEBUG_PORT"
    echo ""

    # Enable security skip mode if setup mode is enabled
    if [ "$SETUP_MODE" = "true" ]; then
        echo "⚠️  Setup mode enabled - Starting with security disabled temporarily"
        echo ""
        export THUNDER_SKIP_SECURITY=true
    fi

    # Run debugger
    dlv exec --listen=:$DEBUG_PORT --headless=true --api-version=2 --accept-multiclient --continue ./thunder &
    THUNDER_PID=$!
else
    echo "⚡ Starting Thunder Server ..."

    # Enable security skip mode if setup mode is enabled
    if [ "$SETUP_MODE" = "true" ]; then
        echo "⚠️  Setup mode enabled - Starting with security disabled temporarily"
        echo ""
        export THUNDER_SKIP_SECURITY=true
    fi

    BACKEND_PORT=$BACKEND_PORT ./thunder &
    THUNDER_PID=$!
fi

# Cleanup function
cleanup() {
    echo -e "\n🛑 Stopping server..."
    if [ -n "$THUNDER_PID" ]; then
        kill $THUNDER_PID 2>/dev/null || true
    fi
}

# Cleanup on Ctrl+C
trap cleanup SIGINT

# Run initial setup if requested
if [ "$SETUP_MODE" = "true" ]; then
    echo "⚙️  Running initial data setup..."
    echo ""

    # Run the setup script - it will handle server readiness checking
    ./scripts/setup_initial_data.sh --port "$BACKEND_PORT"

    if [ $? -eq 0 ]; then
        echo ""
        echo "✅ Initial data setup completed successfully"
        echo ""
        echo "⚠️  Server is still running with SECURITY DISABLED"
        echo ""
        echo "💡 To enable security:"
        echo "   1. Stop the server (Ctrl+C)"
        echo "   2. Restart without --setup flag: ./start.sh"
        echo ""
    else
        echo ""
        echo "❌ Initial data setup failed"
        echo "💡 Check the logs above for more details"
        echo "💡 You can run the setup manually using: ./scripts/setup_initial_data.sh --port $BACKEND_PORT"
        echo ""
        echo "⚠️  Server is still running with security disabled"
    fi
fi

# Status
echo ""
echo "🚀 Server running"
echo "Press Ctrl+C to stop the server."

# Wait for background processes
wait $THUNDER_PID
