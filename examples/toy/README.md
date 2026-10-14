## examples/toy

This is a tiny Goa design, with smoke tests, used to **doc-test** the goa-vcr plugin.

It is a Go module of its own (`github.com/crossnokaye/goa-vcr/examples/toy`, with a `replace` pointing at the repository root): the smoke tests import generated packages that are not committed, and keeping them out of the root module keeps `go vet ./...` green on a fresh checkout.

### What it proves

- `go run goa.design/goa/v3/cmd/goa@v3.23.4 gen ...` generates normal Goa `gen/` + `gen/http/` output **plus** `gen/http/toy/vcr` (because the design blank-imports the plugin).
- The generated VCR glue **compiles** and the key playback rules **work**:
  - unary scenario handler optional (fallback to stub-backed background)
  - loopback header bypasses unary scenarios
  - streaming scenario handler required (SSE example)

### How to run it locally

From this directory:

```bash
# Use a Goa CLI version compatible with the repo (this repo pins goa to v3.23.4).
go run goa.design/goa/v3/cmd/goa@v3.23.4 gen github.com/crossnokaye/goa-vcr/examples/toy/design -o .
go test ./...
```

Or let the Go test do all of it, from the repository root:

```bash
go test ./plugin/vcr
```
