## 1. Update Go version references

- [ ] 1.1 Update `go.mod` directive from `go 1.21` to `go 1.24`
- [ ] 1.2 Update Dockerfile builder image from `golang:1.21-alpine3.18` to `golang:1.24-alpine3.18`
- [ ] 1.3 Update `.github/workflows/ci.yaml` go-version from `"1.21"` to `"1.24"` and step name to `Set up Go 1.24`

## 2. Validate and fix

- [ ] 2.1 Run `go mod tidy` to update `go.sum` and accept any `toolchain` directive
- [ ] 2.2 Run `go build ./...` and fix any compilation errors
- [ ] 2.3 Run `go vet ./...` and fix any new vet findings
- [ ] 2.4 Run `go test ./...` and verify all tests pass
