# Minimal Debian Bookworm-based Development Container for Go
# Base: Debian 12 (Bookworm) Slim - stable LTS with glibc compatibility
FROM debian:bookworm-slim

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Set default user and paths
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# ==============================================================================
# SYSTEM ESSENTIALS
# ==============================================================================
# Install absolutely minimal system packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Core utilities
    ca-certificates \
    curl \
    git \
    # Process tools (ps, top, etc.)
    procps \
    # User/group management
    sudo \
    # Clean up apt cache to reduce image size
    && rm -rf /var/lib/apt/lists/*

# ==============================================================================
# DEVELOPMENT UTILITIES
# ==============================================================================

# Install CLI tools in one layer
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Text processing
    jq \
    less \
    # Search tools
    ripgrep \
    # File navigation
    tree \
    # Decompression (needed for binary installs below)
    unzip \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

# Install GitHub CLI from official apt repository
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/* \
    && gh --version

# ==============================================================================
# LANGUAGE RUNTIMES
# ==============================================================================

# Node.js 22.x (required for AI CLI tools)
# Codex CLI requires Node 22+, Claude Code requires Node 18+
# Install via NodeSource for current LTS version
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# Verify Node.js installation
RUN node --version && npm --version

# ==============================================================================
# AI/DEVELOPMENT TOOLS
# ==============================================================================

# Install AI coding assistant CLIs and development tools globally
# Combine into single layer and clean npm cache to save space
RUN npm install -g \
    @anthropic-ai/claude-code \
    @openai/codex \
    @github/copilot \
    @google/gemini-cli \
    @charmland/crush \
    wrangler \
    && npm cache clean --force \
    && claude --version || echo "Claude Code installed" \
    && codex --version || echo "Codex installed" \
    && copilot --version || echo "GitHub Copilot installed" \
    && gemini --version || echo "Gemini CLI installed" \
    && crush --version || echo "Crush installed" \
    && wrangler --version || echo "Wrangler installed"

# Install OpenCode (Go binary) - download directly from GitHub releases
RUN ARCH=$(dpkg --print-architecture) \
    && if [ "$ARCH" = "arm64" ]; then \
        OPENCODE_ARCH="arm64"; \
    else \
        OPENCODE_ARCH="x86_64"; \
    fi \
    && OPENCODE_VERSION="0.0.55" \
    && mkdir -p /tmp/opencode-install \
    && cd /tmp/opencode-install \
    && curl -fsSL "https://github.com/opencode-ai/opencode/releases/download/v${OPENCODE_VERSION}/opencode-linux-${OPENCODE_ARCH}.tar.gz" -o opencode.tar.gz \
    && tar -xzf opencode.tar.gz \
    && mv opencode /usr/local/bin/ \
    && chmod +x /usr/local/bin/opencode \
    && cd / \
    && rm -rf /tmp/opencode-install \
    && opencode --version || echo "OpenCode installed"

# ==============================================================================
# USER SETUP
# ==============================================================================
# Create non-root user with sudo access (matches VS Code expectations)
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# ==============================================================================
# GO INSTALLATION
# ==============================================================================
# Install Go from official tarball
ENV GO_VERSION=1.23.5
ENV PATH=/usr/local/go/bin:$PATH

RUN ARCH=$(dpkg --print-architecture) \
    && if [ "$ARCH" = "arm64" ]; then \
        GO_ARCH="arm64"; \
    else \
        GO_ARCH="amd64"; \
    fi \
    # Go - use temp dir for extraction
    && mkdir -p /tmp/go-install \
    && cd /tmp/go-install \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o go.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && cd / \
    && rm -rf /tmp/go-install \
    # Verify installation
    && go version

# ==============================================================================
# WORKSPACE SETUP
# ==============================================================================
WORKDIR /workspace

# Switch to non-root user
USER $USERNAME

# Default shell
CMD ["/bin/bash"]
