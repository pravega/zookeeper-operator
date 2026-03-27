## Why

The project uses Go 1.21 which is past end-of-life. Updating to Go 1.24 brings security patches, performance improvements, and ensures compatibility with modern tooling. This is a minimal version bump — dependency upgrades are out of scope.

## What Changes

- Update Go version directive in `go.mod` from `1.21` to `1.24`
- Update Go builder image in `Dockerfile` from `golang:1.21-alpine3.18` to `golang:1.24-alpine3.18`
- Update Go version in `.github/workflows/ci.yaml` from `1.21` to `1.24`
- Resolve any compilation issues introduced by new Go version (e.g. new `vet` checks, language changes)

## Capabilities

### New Capabilities
- `go-version-alignment`: Ensure all Go version references across the project are consistent and set to Go 1.24

### Modified Capabilities
<!-- None — no existing specs to modify -->

## Impact

- **Build files**: `go.mod`, `Dockerfile`, `.github/workflows/ci.yaml`
- **Compilation**: New `go vet` checks or deprecations in Go 1.22–1.24 may surface warnings or errors that need fixing
- **Dependencies**: Existing dependencies should remain compatible (no dependency upgrades in scope), but `go.sum` may be updated by the toolchain
- **CI**: GitHub Actions workflow will use Go 1.24 for builds and tests
