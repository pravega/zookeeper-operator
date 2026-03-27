## Context

The project currently targets Go 1.21 across three locations: `go.mod`, `Dockerfile`, and `.github/workflows/ci.yaml`. The Dockerfile uses `golang:1.21-alpine3.18` as the builder stage. The CI workflow pins `go-version: "1.21"` via `actions/setup-go@v2`.

## Goals / Non-Goals

**Goals:**
- Update all Go version references to 1.24
- Ensure the project compiles and tests pass on Go 1.24

**Non-Goals:**
- Upgrading Go module dependencies (k8s.io, controller-runtime, etc.)
- Updating CI action versions (setup-go, checkout, codecov)
- Updating Alpine or distroless base image versions
- Adopting new Go 1.22–1.24 language features

## Decisions

**1. Update `go.mod` directive to `go 1.24`**
Go 1.24 may add a `toolchain` directive automatically when running `go mod tidy`. We accept this as expected behavior. Alternative: pin `go 1.22` as a safer intermediate step — rejected because 1.24 is the stated target and there's no benefit to a staged approach for a version directive change.

**2. Keep Alpine version at 3.18 in Dockerfile**
The `golang:1.24-alpine3.18` image should exist. If it doesn't, we'll use the latest Alpine tag available for `golang:1.24`. Alpine version bumps are out of scope.

**3. Fix compilation issues reactively**
Rather than preemptively auditing all Go 1.22–1.24 changes, we'll update the version, build, and fix any issues that surface. This is the most efficient approach for a small codebase.

## Risks / Trade-offs

- **[Risk] `golang:1.24-alpine3.18` image may not exist** → Fall back to `golang:1.24-alpine` (latest Alpine for that Go version)
- **[Risk] New `go vet` or `go fmt` checks may cause CI failures** → Fix any surfaced issues as part of this change
- **[Risk] `go mod tidy` may modify `go.sum` or add `toolchain` directive** → Accept these as expected side effects of the version bump
