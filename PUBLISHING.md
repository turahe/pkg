# Publishing Guide

This guide explains how to publish this Go module to make it available via `go get`.

## Prerequisites

1. A GitHub account (or GitLab, Bitbucket, etc.)
2. Git installed on your machine
3. The repository should be accessible via the module path in `go.mod`

## Step 1: Initialize Git Repository

```bash
# Initialize git repository
git init

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: Go package collection"
```

## Step 2: Create GitHub Repository

1. Go to [GitHub](https://github.com) and create a new repository
2. Repository name should match your module path: `pkg`
3. Organization/user should match: `turahe`
4. **Do NOT** initialize with README, .gitignore, or license (we already have these)
5. Copy the repository URL (e.g., `https://github.com/turahe/pkg.git`)

## Step 3: Connect Local Repository to GitHub

```bash
# Add remote origin (replace with your actual repository URL)
git remote add origin https://github.com/turahe/pkg.git

# Verify remote
git remote -v

# Push to GitHub
git branch -M main
git push -u origin main
```

## Step 4: Create Version Tag

Go modules use semantic versioning (v0.0.1, v1.0.0, etc.). Create your first release:

```bash
# Create and push version tag
git tag v0.1.0
git push origin v0.1.0

# Or create annotated tag (recommended)
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
```

### Versioning Guidelines

- **v0.x.x** - Initial development, API may change
- **v1.x.x** - Stable API, backward compatible
- **v2.x.x+** - Major version changes, may require module path changes

## Step 5: Verify Module Works

After publishing, verify the module can be downloaded:

```bash
# Test in a new directory
cd /tmp
mkdir test-module
cd test-module

# Initialize a test Go module
go mod init test

# Try to get your module
go get github.com/turahe/pkg@v0.1.0

# Verify it's in go.mod
cat go.mod
```

## Step 6: Update Module Path (if needed)

If your GitHub username/organization doesn't match `turahe`, update `go.mod`:

```bash
# Update module path in go.mod
# Change: module github.com/turahe/pkg
# To:     module github.com/YOUR-USERNAME/pkg
```

Then update all import statements in your code to match.

## Step 7: Create GitHub Release (Optional but Recommended)

1. Go to your repository on GitHub
2. Click "Releases" → "Create a new release"
3. Choose the tag you created (e.g., `v0.1.0`)
4. Add release title and description
5. Click "Publish release"

## Step 8: Future Updates

For future versions:

```bash
# Make your changes
git add .
git commit -m "Add new feature"

# Push changes
git push

# Create new version tag
git tag -a v0.2.0 -m "Add new features"
git push origin v0.2.0
```

## Using the Published Module

Once published, others can use your module:

```bash
# Get latest version
go get github.com/turahe/pkg

# Get specific version
go get github.com/turahe/pkg@v0.1.0

# Get latest v0.x.x
go get github.com/turahe/pkg@v0

# Update to latest
go get -u github.com/turahe/pkg
```

## Troubleshooting

### Module not found
- Ensure the repository is public (or you have access)
- Verify the module path in `go.mod` matches the repository URL
- Check that you've pushed tags: `git push --tags`

### Version not found
- Make sure you've created and pushed the tag
- Tags must follow semantic versioning (v0.1.0, not 0.1.0)
- Wait a few minutes for GitHub/Go proxy to index

### Import errors
- Verify all import paths use the correct module path
- Run `go mod tidy` to clean up dependencies

## Additional Resources

- [Go Modules Reference](https://go.dev/ref/mod)
- [Publishing Go Modules](https://go.dev/doc/modules/publishing)
- [Semantic Versioning](https://semver.org/)
