# GitHub Actions Workflows

This directory contains the CI/CD workflows for the SpaceNet project.

## Workflows

### `CI.yml`
Main continuous integration workflow that:
- Tests both server and TUI components
- Lints both server and TUI code with golangci-lint
- Builds binaries for multiple platforms (Linux, macOS, Windows)
- Uploads coverage reports and build artifacts
- Runs on every push and pull request

### `docker-server.yml`
Docker build and publish workflow that:
- Builds multi-architecture Docker images (amd64, arm64)
- Publishes to GitHub Container Registry (ghcr.io)
- Runs security scans with Trivy
- Tests the built Docker image
- Generates build provenance attestations
- Triggers on:
  - Push to main branch (when server code changes)
  - Tags starting with 'v' (for releases)
  - Pull requests (build only, no publish)
  - Manual workflow dispatch


## Image Tags

The Docker workflow publishes images with multiple tags:

- `main` - Latest main branch build
- `latest` - Latest stable release (main branch)
- `edge` - Development builds from main
- `v1.2.3` - Semantic version tags for releases
- `v1.2` - Major.minor version tags
- `v1` - Major version tags
- `pr-123` - Pull request builds (not published)

## Security

- Images are scanned for vulnerabilities using Trivy
- Build provenance is recorded for supply chain security
- Images run as non-root user
- SARIF results are uploaded to GitHub Security tab

## Usage

To use the published Docker image:

```bash
# Pull the latest image
docker pull ghcr.io/bjia56/spacenet/spacenet-server:latest

# Run with default settings
docker run -p 8080:8080 ghcr.io/bjia56/spacenet/spacenet-server:latest

# Run with persistent database
docker run -p 8080:8080 -v $(pwd)/data:/app/data ghcr.io/bjia56/spacenet/spacenet-server:latest
```