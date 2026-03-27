## ADDED Requirements

### Requirement: Go module version directive is 1.24
The `go.mod` file SHALL declare `go 1.24` as the minimum Go version.

#### Scenario: go.mod specifies Go 1.24
- **WHEN** inspecting the `go` directive in `go.mod`
- **THEN** it SHALL read `go 1.24`

### Requirement: Dockerfile uses Go 1.24 builder image
The Dockerfile SHALL use a `golang:1.24`-based Alpine image for the build stage.

#### Scenario: Builder image references Go 1.24
- **WHEN** building the Docker image
- **THEN** the builder stage SHALL use `golang:1.24-alpine*` as the base image

### Requirement: CI workflow uses Go 1.24
The GitHub Actions CI workflow SHALL configure Go 1.24 for builds and tests.

#### Scenario: CI installs Go 1.24
- **WHEN** the CI workflow runs the `setup-go` step
- **THEN** `go-version` SHALL be set to `"1.24"`

#### Scenario: CI step name reflects version
- **WHEN** the CI workflow runs
- **THEN** the setup step name SHALL reference Go 1.24

### Requirement: Project compiles on Go 1.24
The project SHALL compile without errors using Go 1.24.

#### Scenario: Successful build
- **WHEN** running `go build ./...` with Go 1.24
- **THEN** the build SHALL succeed with exit code 0

#### Scenario: Tests pass
- **WHEN** running `go test ./...` with Go 1.24
- **THEN** all existing tests SHALL pass

### Requirement: Go vet passes on Go 1.24
The project SHALL pass `go vet` checks under Go 1.24.

#### Scenario: Vet clean
- **WHEN** running `go vet ./...` with Go 1.24
- **THEN** no vet errors SHALL be reported
