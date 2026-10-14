### Testing Your Changes

To see sample generation results, gen the toy example. It is a nested Go module, so run from its directory:

```bash
cd examples/toy
go run goa.design/goa/v3/cmd/goa@v3.23.4 gen github.com/crossnokaye/goa-vcr/examples/toy/design -o .
```

The result appears under `examples/toy/gen` which is gitignored.

Tests:

```bash
go vet ./... && go test ./...   # generator and runtime tests, and the toy pipeline
```

- `go test ./...` runs the toy pipeline (gen -> `go test ./...` in `examples/toy`). `go test -short ./...` skips it.
- The smoke tests for the generated VCR glue live in `examples/toy/*_test.go`. They import `examples/toy/gen`, so they only compile after generation.
