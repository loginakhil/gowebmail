# Release Guide

This guide explains how to create releases for GoWebMail using GoReleaser and GitHub Actions.

## Overview

The project uses [GoReleaser](https://goreleaser.com/) to automate the release process, which includes:

- Building binaries for multiple platforms (Linux, macOS, Windows)
- Creating Docker images for multiple architectures (amd64, arm64)
- Generating checksums
- Creating GitHub releases with changelogs
- Publishing Docker images to Docker Hub

## Prerequisites

1. **GitHub Secrets** configured (see [DOCKER_HUB_SETUP.md](DOCKER_HUB_SETUP.md)):
   - `DOCKERHUB_USERNAME`
   - `DOCKERHUB_TOKEN`

2. **Repository permissions**:
   - Write access to create tags
   - Admin access to configure secrets

## Creating a Release

### 1. Prepare the Release

Before creating a release, ensure:

- All changes are merged to the main branch
- Tests are passing
- Documentation is updated
- CHANGELOG is ready (optional, auto-generated from commits)

### 2. Create and Push a Version Tag

GoReleaser is triggered by pushing a tag that matches the pattern `v*.*.*` (semantic versioning).

```bash
# Ensure you're on the main branch and up to date
git checkout main
git pull origin main

# Create a new tag (replace with your version)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

### 3. Automated Release Process

Once the tag is pushed, GitHub Actions will automatically:

1. **Build binaries** for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

2. **Create archives** containing:
   - Binary executable
   - README.md
   - LICENSE
   - Example configuration file
   - Web assets

3. **Build Docker images** for:
   - linux/amd64
   - linux/arm64

4. **Push to Docker Hub** with tags:
   - `v1.0.0` (specific version)
   - `latest` (latest release)

5. **Create GitHub Release** with:
   - Auto-generated changelog
   - Binary downloads
   - Checksums
   - Installation instructions

### 4. Monitor the Release

1. Go to the **Actions** tab in your GitHub repository
2. Click on the **Release** workflow
3. Monitor the progress of the release build
4. Check for any errors

### 5. Verify the Release

Once complete, verify:

1. **GitHub Release**: Check the [Releases page](../../releases)
   - Release notes are generated
   - All binary assets are attached
   - Checksums file is present

2. **Docker Hub**: Verify images are published
   ```bash
   docker pull <your-username>/gowebmail:v1.0.0
   docker pull <your-username>/gowebmail:latest
   ```

3. **Binary Downloads**: Test downloading and running a binary
   ```bash
   # Example for Linux
   wget https://github.com/<owner>/gowebmail/releases/download/v1.0.0/gowebmail_1.0.0_Linux_x86_64.tar.gz
   tar -xzf gowebmail_1.0.0_Linux_x86_64.tar.gz
   ./gowebmail --version
   ```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v2.0.0): Incompatible API changes
- **MINOR** version (v1.1.0): New functionality, backwards compatible
- **PATCH** version (v1.0.1): Backwards compatible bug fixes

Examples:
- `v1.0.0` - First stable release
- `v1.1.0` - Added new features
- `v1.1.1` - Bug fixes
- `v2.0.0` - Breaking changes

## Pre-releases

For beta or release candidate versions:

```bash
git tag -a v1.0.0-beta.1 -m "Release v1.0.0-beta.1"
git push origin v1.0.0-beta.1
```

GoReleaser will automatically mark these as pre-releases on GitHub.

## Changelog Generation

The changelog is automatically generated from commit messages. For best results, use conventional commits:

- `feat: Add new feature` → Features section
- `fix: Fix bug` → Bug Fixes section
- `perf: Improve performance` → Performance Improvements section
- `docs: Update documentation` → Excluded from changelog
- `chore: Update dependencies` → Excluded from changelog

Example commit messages:
```bash
git commit -m "feat: Add email search functionality"
git commit -m "fix: Resolve SQLite FTS5 compatibility issue"
git commit -m "perf: Optimize database queries"
```

## Local Testing

To test the release process locally without publishing:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Run a snapshot build (doesn't publish)
goreleaser release --snapshot --clean

# Check the dist/ directory for generated artifacts
ls -la dist/
```

## Troubleshooting

### Release Workflow Fails

**Problem**: Build fails with authentication error

**Solution**: 
- Verify `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets are set correctly
- Ensure the Docker Hub token hasn't expired

**Problem**: CGO compilation errors

**Solution**:
- The workflow installs cross-compilation tools automatically
- Check the workflow logs for specific compilation errors
- Some platforms may require additional build tags or flags

### Docker Images Not Published

**Problem**: Images build but don't appear on Docker Hub

**Solution**:
- Check that `skip_push: false` in `.goreleaser.yml`
- Verify Docker Hub credentials are correct
- Ensure you have push permissions to the repository

### Missing Assets in Release

**Problem**: Some files are missing from the release archives

**Solution**:
- Check the `files` section in `.goreleaser.yml`
- Ensure the files exist in the repository
- Verify the paths are correct

## Manual Release (Emergency)

If the automated process fails, you can create a release manually:

```bash
# Build locally
goreleaser release --clean

# This will:
# - Build all binaries
# - Create archives
# - Generate checksums
# - Create a GitHub release
# - Push Docker images
```

Note: You'll need to set environment variables:
```bash
export GITHUB_TOKEN="your-github-token"
export DOCKERHUB_USERNAME="your-username"
```

## Configuration Files

- **`.goreleaser.yml`**: Main GoReleaser configuration
- **`.github/workflows/release.yml`**: GitHub Actions workflow for releases
- **`.github/workflows/docker-publish.yml`**: Continuous Docker builds (non-release)

## Support

For issues with the release process:

1. Check the [GoReleaser documentation](https://goreleaser.com/)
2. Review GitHub Actions logs
3. Open an issue in the repository
