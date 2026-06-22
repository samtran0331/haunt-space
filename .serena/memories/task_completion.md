# Task Completion Checklist

Run after any coding change:

1. `go vet ./...` — catch obvious errors
2. `go test ./...` — run all tests
3. `go build -o hsp .` — verify binary compiles
4. `gofmt -w .` — format (or verify no diff)

For wizard/TUI changes: manual smoke test with `go run . wizard` since tests can't cover interactive TUI.
