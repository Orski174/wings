# GitHub Secrets Quick Reference

## ✅ Required Secrets: NONE

All workflows use GitHub's automatically provided `GITHUB_TOKEN`. No manual secret configuration needed!

## Secrets Summary

| Secret | Required? | Source | Usage |
|--------|-----------|--------|-------|
| `GITHUB_TOKEN` | ✅ Yes | Auto-provided by GitHub | Release creation, Docker registry auth |

## What This Means

✅ **No Action Required** - All workflows work out of the box  
✅ **No Manual Setup** - GitHub provides all necessary credentials  
✅ **Works on Forks** - Automatic tokens work in any repository  

## Setup Checklist for New Repositories/Forks

- [ ] **Verify Workflow Permissions**
  - Go to: Settings → Actions → General → Workflow permissions
  - Select: "Read and write permissions"
  - Check: "Allow GitHub Actions to create and approve pull requests"
  
- [ ] **Verify GitHub Packages Enabled**
  - Should be enabled by default for public repos
  - Docker images will publish to: `ghcr.io/YOUR_USERNAME/wings`

## Workflow Status

| Workflow | Secrets Needed | Auto-Configured |
|----------|----------------|-----------------|
| release.yaml | GITHUB_TOKEN | ✅ Yes |
| docker.yaml | GITHUB_TOKEN | ✅ Yes |
| codeql.yaml | None | ✅ Yes |
| push.yaml | None | ✅ Yes |

## Common Issues

### Release fails with "Resource not accessible"
**Fix:** Enable "Read and write permissions" in Settings → Actions → General

### Docker push fails with "permission denied"
**Fix:** Verify GitHub Container Registry access and workflow has `packages: write` permission

## Full Documentation

See [SECRETS_GUIDE.md](.github/SECRETS_GUIDE.md) for complete details.
