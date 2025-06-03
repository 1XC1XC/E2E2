#!/bin/bash

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored text
print_color() {
    printf "${!1}%s${NC}\n" "$2"
}

# Flag to avoid running cleanup multiple times
CLEANUP_DONE=false

# Function to clean up and exit
cleanup() {
    if [ "$CLEANUP_DONE" = false ]; then
        CLEANUP_DONE=true
        print_color "YELLOW" "Cleaning up..."
        if [ ! -z "$SERVER_PID" ]; then
            kill $SERVER_PID 2>/dev/null
            wait $SERVER_PID 2>/dev/null
            print_color "GREEN" "Server stopped."
        fi
    fi
    exit 0
}

# Set up trap to call cleanup function on script exit or interrupt
trap cleanup EXIT SIGINT

# Start the server
print_color "GREEN" "Starting the server..."
cd Server
go run . &
SERVER_PID=$!
cd ..

# Wait for server to start
sleep 2

if ps -p $SERVER_PID >/dev/null; then
    print_color "GREEN" "Server started successfully (PID: $SERVER_PID)."
else
    print_color "RED" "Failed to start the server."
    exit 1
fi

# Main loop
first_run=true
while true; do
    if [ "$first_run" = true ]; then
        first_run=false
    else
        # Line separation after first run
        echo -e "\n${BLUE}--------------------------------------------${NC}\n"
    fi

    print_color "YELLOW" "Press Enter to run the client, or Ctrl+C to stop the server and exit."
    read -r

    print_color "YELLOW" "Running the client..."
    cd Client
    go run .
    cd ..
done
