# Development Container Configuration

This directory contains a minimal devcontainer setup for Go development, based on Debian 12 (Bookworm) Slim.

## Philosophy

- **Minimal by default**: Only Go and essential tools
- **Transparent**: Know exactly what's in your container
- **glibc-based**: Full compatibility with pre-compiled binaries
- **Well-documented**: Clear sections for customization

## Structure

```
.devcontainer/
├── Dockerfile           # Container image definition
├── devcontainer.json    # VS Code configuration
└── README.md           # This file
```

## Base Image

**`debian:bookworm-slim`** (~75MB base)
- Debian 12 LTS (stable until ~2028)
- Uses glibc (not musl like Alpine)
- Official Debian minimal image
- APT package manager

## What's Included

The container includes essential tools for Go development:

**System essentials:**
- `ca-certificates` - SSL/TLS certificates
- `curl` - HTTP client
- `git` - Version control
- `procps` - Process utilities (ps, top, etc.)
- `sudo` - Privilege escalation
- Non-root user (`vscode`) with sudo access

**Language runtime:**
- Go 1.23.5 (latest stable)

**Development tools:**
- GitHub CLI (`gh`)
- ripgrep (`rg`)
- jq (JSON processor)
- tree (directory visualization)

## Usage

### Using VS Code

1. Install the "Dev Containers" extension in VS Code
2. Open this project in VS Code
3. Press `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux)
4. Select "Dev Containers: Reopen in Container"

### Rebuild Container

After modifying `Dockerfile` or `devcontainer.json`:
- `Cmd+Shift+P` → "Dev Containers: Rebuild Container"

### Using Without VS Code

Build manually:
```bash
docker build -t kyletube-dev .devcontainer
docker run -it -v $(pwd):/workspace kyletube-dev
```

## Persisted Data

By default, nothing is persisted or mounted from your host machine. This keeps the container portable and not tied to any specific user.

If you want to persist credentials or mount your git config, you can uncomment the mounts section in `devcontainer.json`:
- **SSH keys** (read-only): `~/.ssh` for git operations
- **Git config** (read-only): `~/.gitconfig`
- **Claude credentials** (Docker volume): `kyletube-claude-credentials`

## Environment Variables

The following environment variables are passed through from your host:
- `ANTHROPIC_API_KEY`
- `OPENAI_API_KEY`
- `GITHUB_TOKEN`
- `KAGI_API_KEY`
- `SSH_AUTH_SOCK`

## Customization

### Adding Build Tools

If you need CGO or native compilation:
```dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*
```

### Adding Database Clients

```dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-client \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*
```

## Troubleshooting

### Permission issues
Ensure `remoteUser` in `devcontainer.json` matches the user in `Dockerfile` (default: `vscode`).

### Go module downloads failing
Check that git is working properly and SSH keys are mounted correctly.

## Resources

- [Dev Containers Documentation](https://code.visualstudio.com/docs/devcontainers/containers)
- [Go Documentation](https://go.dev/doc/)
- [Debian Packages Search](https://packages.debian.org/bookworm/)
