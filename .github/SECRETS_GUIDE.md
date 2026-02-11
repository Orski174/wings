# GitHub Workflows - Required Secrets Guide

## Overview

This document provides a complete list of secrets required by the GitHub workflows in this repository, along with detailed instructions on how to obtain and configure them.

## Summary of Required Secrets

| Secret Name | Used In | Auto-Provided | Manual Setup Required |
|------------|---------|---------------|----------------------|
| `GITHUB_TOKEN` | release.yaml, docker.yaml | ✅ Yes | ❌ No |

## Detailed Analysis

### 1. `GITHUB_TOKEN` (Auto-Provided)

**Used In:**
- `.github/workflows/release.yaml` (line 50)
- `.github/workflows/docker.yaml` (line 47)

**Purpose:**
- **release.yaml**: Used by `softprops/action-gh-release` to create GitHub releases with binaries
- **docker.yaml**: Used to authenticate with GitHub Container Registry (ghcr.io) for pushing Docker images

**How to Obtain:**
✅ **AUTOMATICALLY PROVIDED** - GitHub automatically creates and provides this token for every workflow run. No manual setup required.

**Permissions:**
The workflows explicitly define the required permissions:
- **release.yaml**: `contents: write` (to create releases)
- **docker.yaml**: `contents: read`, `packages: write` (to push Docker images to GHCR)

**No Action Required:** This token is managed by GitHub Actions automatically.

---

## Workflow-by-Workflow Breakdown

### Release Workflow (`.github/workflows/release.yaml`)

**Trigger:** Push to tags matching `v*` pattern

**Secrets Used:**
- ✅ `GITHUB_TOKEN` (auto-provided)

**Actions:**
1. Builds release binaries for amd64 and arm64
2. Creates a release branch
3. Updates version in `system/const.go`
4. Creates a draft GitHub release with binaries attached
5. Extracts changelog for the release

**Required Permissions:**
- `contents: write` - To create releases and push to release branch

---

### Docker Workflow (`.github/workflows/docker.yaml`)

**Triggers:**
- Push to `develop` branch
- Release published

**Secrets Used:**
- ✅ `GITHUB_TOKEN` (auto-provided)

**Actions:**
1. Builds multi-platform Docker images (linux/amd64, linux/arm64)
2. Pushes to GitHub Container Registry (ghcr.io)
3. Tags images appropriately:
   - `latest` for stable releases
   - `develop` for development builds
   - Version tags for releases

**Required Permissions:**
- `contents: read` - To read repository code
- `packages: write` - To push to GitHub Container Registry

**Authentication:**
Uses `docker/login-action` with:
- Registry: `ghcr.io`
- Username: `${{ github.actor }}` (automatic)
- Password: `${{ secrets.GITHUB_TOKEN }}` (automatic)

---

### CodeQL Workflow (`.github/workflows/codeql.yaml`)

**Triggers:**
- Push to `develop` branch
- Pull requests to `develop` branch
- Scheduled (Thursdays at 9 AM UTC)

**Secrets Used:**
- ❌ None - Uses default GITHUB_TOKEN permissions

**Actions:**
1. Performs security analysis on Go code
2. Uploads security findings to GitHub Security tab

**Required Permissions:**
- `actions: read`
- `contents: read`
- `security-events: write`

**No Secrets Required**

---

### Push Workflow (`.github/workflows/push.yaml`)

**Triggers:**
- Push to `develop` branch
- Pull requests to `develop` branch

**Secrets Used:**
- ❌ None

**Actions:**
1. Builds Wings for multiple Go versions and architectures
2. Runs tests (including race detection tests)
3. Uploads build artifacts

**Required Permissions:**
- `contents: read`

**No Secrets Required**

---

## Repository Secrets Configuration

### Current Status: ✅ All Required Secrets Are Auto-Provided

**No manual secret configuration is needed!** All workflows use GitHub's automatically provided `GITHUB_TOKEN`, which is available in every workflow run.

### If You Fork This Repository

If you fork this repository, the workflows will continue to work automatically because:

1. **GITHUB_TOKEN** is provided by GitHub Actions in any repository
2. **GitHub Container Registry (GHCR)** authentication uses the automatic token
3. No third-party service tokens are required

**What You Need to Verify:**

1. **Repository Permissions:**
   - Go to repository Settings → Actions → General
   - Under "Workflow permissions", ensure:
     - ✅ "Read and write permissions" is selected (for release workflow)
     - ✅ "Allow GitHub Actions to create and approve pull requests" is checked (if needed)

2. **GitHub Packages:**
   - Packages will be published to `ghcr.io/YOUR_USERNAME/wings`
   - Ensure you have GitHub Packages enabled (it's free for public repos)

---

## Optional Enhancements (Not Currently Used)

The following secrets could be added for additional functionality, but are **NOT required** for current workflows:

### 1. Docker Hub Authentication (Optional)

If you want to also push to Docker Hub:
- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

**How to obtain:**
1. Create account at https://hub.docker.com
2. Go to Account Settings → Security
3. Create a new access token
4. Add to repository secrets

### 2. Third-Party Registry Tokens (Optional)

For other container registries (Quay.io, etc.):
- Registry-specific authentication tokens

### 3. Notification Tokens (Optional)

For Discord/Slack notifications on release:
- `DISCORD_WEBHOOK_URL`
- `SLACK_WEBHOOK_URL`

---

## Troubleshooting

### Issue: Release Creation Fails

**Error:** `Resource not accessible by integration`

**Solution:** Check repository workflow permissions:
1. Go to Settings → Actions → General
2. Under "Workflow permissions", select "Read and write permissions"
3. Save changes

### Issue: Docker Push Fails

**Error:** `denied: permission_denied`

**Solution:**
1. Verify GitHub Container Registry is enabled
2. Check that workflow has `packages: write` permission
3. Ensure `GITHUB_TOKEN` has not expired (it shouldn't, as it's auto-generated)

### Issue: Artifacts Not Uploading

**Error:** `Unable to upload artifact`

**Solution:**
1. Check repository storage limits
2. Verify workflow has `actions: write` permission if needed
3. Check artifact size limits (GitHub has size restrictions)

---

## Summary

### ✅ What's Already Configured

All workflows are production-ready with no manual secret setup required:
- ✅ Release workflow uses auto-provided GITHUB_TOKEN
- ✅ Docker workflow uses auto-provided GITHUB_TOKEN for GHCR
- ✅ CodeQL workflow uses built-in security permissions
- ✅ Push workflow requires no authentication

### ⚠️ What to Check

1. Repository workflow permissions (Settings → Actions → General)
2. Ensure "Read and write permissions" is enabled
3. Verify GitHub Packages/Container Registry is accessible

### 🎯 Required Actions

**For Repository Owner:** None - workflows are ready to use!

**For Fork Maintainers:**
1. Verify workflow permissions in repository settings
2. Packages will publish to your GHCR namespace automatically
3. No manual secrets needed

---

## Additional Resources

- [GitHub Actions Automatic Token](https://docs.github.com/en/actions/security-guides/automatic-token-authentication)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Workflow Permissions](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#permissions)
- [Managing Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)

---

**Last Updated:** February 11, 2026  
**Workflows Analyzed:** release.yaml, docker.yaml, codeql.yaml, push.yaml
