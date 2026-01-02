# Docker Hub Setup Guide

This guide explains how to configure GitHub Actions to automatically build and push Docker images to Docker Hub.

## Prerequisites

1. A Docker Hub account
2. Admin access to this GitHub repository

## Setup Steps

### 1. Create Docker Hub Access Token

1. Log in to [Docker Hub](https://hub.docker.com/)
2. Click on your username in the top right corner
3. Select **Account Settings**
4. Go to **Security** → **Access Tokens**
5. Click **New Access Token**
6. Give it a description (e.g., "GitHub Actions - gowebmail")
7. Set permissions to **Read, Write, Delete**
8. Click **Generate**
9. **Copy the token immediately** (you won't be able to see it again)

### 2. Add Secrets to GitHub Repository

1. Go to your GitHub repository
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add the following secrets:

   **Secret 1: DOCKERHUB_USERNAME**
   - Name: `DOCKERHUB_USERNAME`
   - Value: Your Docker Hub username

   **Secret 2: DOCKERHUB_TOKEN**
   - Name: `DOCKERHUB_TOKEN`
   - Value: The access token you generated in step 1

### 3. Verify Setup

Once the secrets are configured, the workflow will automatically run when:

- Code is pushed to `main` or `master` branch
- A new tag matching `v*.*.*` is created (e.g., `v1.0.0`)
- A pull request is opened (builds only, doesn't push)
- Manually triggered via the Actions tab

## Docker Image Tags

The workflow automatically creates the following tags:

- `latest` - Latest build from the default branch
- `main` or `master` - Latest build from that branch
- `v1.2.3` - Semantic version tags
- `v1.2` - Major.minor version
- `v1` - Major version only
- `main-abc1234` - Branch name with commit SHA

## Multi-Architecture Support

The workflow builds images for:
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM64/Apple Silicon)

## Manual Trigger

To manually trigger a build:

1. Go to the **Actions** tab in your repository
2. Select **Docker Build and Push** workflow
3. Click **Run workflow**
4. Select the branch and click **Run workflow**

## Troubleshooting

### Build Fails with Authentication Error

- Verify that `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` are correctly set
- Ensure the access token has not expired
- Check that the token has the correct permissions

### Image Not Appearing on Docker Hub

- Verify the workflow completed successfully in the Actions tab
- Check that you're not on a pull request (PRs don't push images)
- Ensure the repository name in the workflow matches your Docker Hub repository

### Permission Denied

- Verify your Docker Hub account has permission to push to the repository
- If using an organization, ensure you have the correct role

## Using the Docker Image

Once published, you can pull the image:

```bash
# Pull latest version
docker pull <your-dockerhub-username>/gowebmail:latest

# Pull specific version
docker pull <your-dockerhub-username>/gowebmail:v1.0.0

# Run the container
docker run -p 1025:1025 -p 8080:8080 <your-dockerhub-username>/gowebmail:latest
```

## Updating the Workflow

The workflow file is located at `.github/workflows/docker-publish.yml`. You can customize:

- Trigger conditions
- Build platforms
- Tag naming strategy
- Build arguments
- Cache settings
