# Devcontainer management recipes

# List all devcontainer recipes
default:
    @just --list devcontainer

# Build the devcontainer
build:
    devcontainer build --workspace-folder .

# Rebuild the devcontainer from scratch (no cache)
rebuild: down
    devcontainer build --workspace-folder . --no-cache

# Start the devcontainer (if not already running)
up:
    devcontainer up --workspace-folder .

# Stop and remove the devcontainer
down:
    docker ps -a --filter "label=devcontainer.local_folder=$(pwd)" --format {{'{{.ID}}'}} | xargs -r docker rm -f

# Drop into a bash shell in the devcontainer (starts container if not running)
shell:
    #!/usr/bin/env bash
    set -euo pipefail
    # Check if container is running, start it if not
    if ! devcontainer exec --workspace-folder . true 2>/dev/null; then
        echo "Container not running, starting..."
        devcontainer up --workspace-folder .
    fi
    devcontainer exec --workspace-folder . bash

# Run a command in the devcontainer
run CMD:
    devcontainer exec --workspace-folder . {{CMD}}

# Open VS Code in the devcontainer
code:
    code .

# Test the container by running version checks
test:
    @echo "Testing container tools..."
    @devcontainer exec --workspace-folder . bash -c "\
        echo '=== Language Runtimes ===' && \
        echo 'Node.js:' && node --version && \
        echo 'npm:' && npm --version && \
        echo 'Go:' && go version && \
        echo '' && \
        echo '=== AI Coding Assistants ===' && \
        echo 'Claude Code:' && (claude --version || echo 'installed') && \
        echo 'OpenAI Codex:' && (codex --version || echo 'installed') && \
        echo 'GitHub Copilot:' && (copilot --version || echo 'installed') && \
        echo 'Gemini CLI:' && (gemini --version || echo 'installed') && \
        echo 'Crush:' && (crush --version || echo 'installed') && \
        echo 'OpenCode:' && (opencode --version || echo 'installed') && \
        echo '' && \
        echo '=== Development Tools ===' && \
        echo 'git:' && git --version && \
        echo 'GitHub CLI:' && gh --version && \
        echo 'Wrangler:' && (wrangler --version || echo 'installed') && \
        echo 'ripgrep:' && rg --version && \
        echo 'jq:' && jq --version \
    "

# Show running devcontainers
ps:
    @docker ps --filter "label=devcontainer.local_folder" --format {{'table {{.Names}}\t{{.Status}}\t{{.Image}}'}}

# Clean up devcontainer images
clean:
    @docker images --filter "label=devcontainer.local_folder=$(pwd)" --format {{'{{.ID}}'}} | xargs -r docker rmi

# List credential volumes
volumes:
    @docker volume ls | grep kyletube

# Delete credential volumes (logout/reset credentials)
clean-volumes:
    @echo "This will delete your Claude credentials. Are you sure? (Ctrl-C to cancel)"
    @read -p "Press Enter to continue..."
    @docker volume rm kyletube-claude-credentials 2>/dev/null || true
    @echo "Credentials deleted. Run 'claude login' next time you start the container."
