#!/usr/bin/env bash
# dev.sh - Run multiple dev servers
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

CYAN='\033[36m'
GREEN='\033[32m'
RESET='\033[0m'

# Check if tmux is available
if ! command -v tmux &> /dev/null; then
    echo "tmux not found. Running services in background."
    trap 'kill $(jobs -p) 2>/dev/null' EXIT

    for service in "$@"; do
        case $service in
            backend)
                echo -e "${CYAN}Starting backend...${RESET}"
                (cd "$PROJECT_ROOT/backend" && air) &
                ;;
            frontend)
                echo -e "${CYAN}Starting frontend...${RESET}"
                (cd "$PROJECT_ROOT/frontend" && npm run dev) &
                ;;
        esac
    done
    wait
    exit 0
fi

SESSION_NAME="homephotos-dev"
tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true
tmux new-session -d -s "$SESSION_NAME" -c "$PROJECT_ROOT"

pane_index=0
for service in "$@"; do
    if [ $pane_index -gt 0 ]; then
        tmux split-window -h -t "$SESSION_NAME" -c "$PROJECT_ROOT"
    fi

    case $service in
        backend)
            tmux send-keys -t "$SESSION_NAME" "cd backend && air" C-m
            ;;
        frontend)
            tmux send-keys -t "$SESSION_NAME" "cd frontend && npm run dev" C-m
            ;;
    esac
    ((pane_index++)) || true
done

tmux select-layout -t "$SESSION_NAME" tiled
echo -e "${GREEN}Dev servers starting in tmux session: $SESSION_NAME${RESET}"
if [ -n "$TMUX" ]; then
    exec tmux switch-client -t "$SESSION_NAME"
else
    exec tmux attach -t "$SESSION_NAME"
fi
