# Docker Build Cross-Platform Compatibility Fix

## Problem Description

The original Docker build was failing on Linux servers (Ubuntu) and GitHub Actions while working locally on Mac with Apple Silicon. This was due to architecture detection issues in the Ignite CLI installation script when used in Docker multi-platform builds.

### Root Cause

The original Dockerfile used:
```dockerfile
RUN curl -L "https://get.ignite.com/cli@${IGNITE_VERSION}!" | bash
```

This installation script (`get.ignite.com/cli`) has limitations in Docker Buildx multi-platform builds:
- May not correctly detect the target architecture during cross-compilation
- Could install binaries in unexpected locations
- Doesn't handle Docker's `TARGETARCH` and `TARGETOS` build arguments properly

## Solution

### 1. Explicit Architecture Handling

The updated Dockerfile now uses Docker's built-in build arguments:
- `TARGETARCH`: Provided by Docker Buildx (amd64, arm64, etc.)
- `TARGETOS`: Provided by Docker Buildx (linux, windows, etc.)

### 2. Direct Binary Download

Instead of relying on the auto-detection script, we now:
- Download the correct architecture-specific binary directly from GitHub releases
- Use explicit architecture matching with a case statement
- Verify the binary works before proceeding

### 3. Updated Dockerfile

```dockerfile
ARG TARGETARCH
ARG TARGETOS

# Set up architecture-specific variables
RUN case ${TARGETARCH} in \
    "amd64") IGNITE_ARCH=amd64 ;; \
    "arm64") IGNITE_ARCH=arm64 ;; \
    *) echo "Unsupported architecture: ${TARGETARCH}" && exit 1 ;; \
    esac && \
    echo "Building for ${TARGETOS}/${TARGETARCH}, downloading ignite-${IGNITE_ARCH}" && \
    curl -L "https://github.com/ignite/cli/releases/download/${IGNITE_VERSION}/ignite-${IGNITE_ARCH}" -o /tmp/ignite && \
    chmod +x /tmp/ignite && \
    mv /tmp/ignite /usr/local/bin/ignite

# Verify the binary works
RUN /usr/local/bin/ignite version || (echo "Ignite binary verification failed" && exit 1)
```

## Testing

### Local Testing
```bash
# Test single platform build
docker build -t arda-poc-test .

# Test multi-platform build (if buildx is available)
docker buildx build --platform linux/amd64,linux/arm64 -t arda-poc-test .

# Test container execution
docker run --rm arda-poc-test arda-pocd version
```

### Automated Testing
Run the provided test script:
```bash
./scripts/test-docker-build.sh
```

## Benefits

1. **Reproducible Builds**: The same Dockerfile now works consistently across different environments
2. **Multi-platform Support**: Proper handling of both `linux/amd64` and `linux/arm64` architectures
3. **Better Error Handling**: Binary verification ensures the download was successful
4. **CI/CD Compatibility**: Works correctly in GitHub Actions and other CI systems
5. **Debugging**: Clear logging of which architecture is being built

## Verification

The fix ensures that:
- ✅ Local development on Mac with Apple Silicon works
- ✅ GitHub Actions CI builds work for both architectures
- ✅ DigitalOcean droplet builds work
- ✅ Any Linux server with Docker can build the image

## Additional Notes

- The fix is backward compatible with existing build processes
- No changes needed to docker-compose files or GitHub workflows
- The Docker Buildx multi-platform feature (`platforms: linux/amd64,linux/arm64`) in the GitHub workflow continues to work as expected

This fix resolves the cross-platform compatibility issues and ensures reproducible Docker builds across different operating systems and architectures.