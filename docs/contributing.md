# Contributing

Contributions are welcome. `go-synctex/synctex` is built to a small set of
non-negotiable rules — they are what keep the parser pure-Go, correct, and safe
to point at untrusted input. Please read these before opening a pull request.

## Hard rules

- **Build from source — no vendoring.** Everything compiles from source, standard
  library only. Do not reach for prebuilt binaries or vendored blobs as a
  shortcut; being able to compile from source is a guarantee of independence.
- **100% test coverage target, enforced in CI.** New code ships with tests, and
  coverage is a CI gate. Fill the error branches — the bad gzip stream, the
  over-cap body, the malformed record line, the out-of-root path — not just the
  happy path.
- **All GitHub content in English.** Issues, pull requests, commits, comments,
  and discussions are English-only.
- **Pure Go, cgo disabled.** The whole point is a single static binary with no C
  toolchain, embeddable anywhere. Code must build with `CGO_ENABLED=0`. If a
  feature seems to need C, it needs a pure-Go path instead.
- **Zero external dependencies.** The module depends only on the Go standard
  library. Keep it that way; a new third-party dependency needs a very good
  reason.
- **Hostile input is the default assumption.** A `.synctex.gz` may come from an
  untrusted build. Keep the decompression cap, keep `WithProjectRoot`'s
  containment checks, and add tests for any new input-handling path.

## Workflow

1. Pick or open an issue describing the change.
2. Work test-first: add the unit tests (happy path **and** error branches), then
   make them pass.
3. Run the full suite with coverage and confirm the gate is green:

    ```sh
    COVERPKG=$(go list ./... | paste -sd, -)
    go test -race -coverpkg="$COVERPKG" -coverprofile=cover.out ./...
    go tool cover -func=cover.out | tail -1   # 100.0%
    ```

4. Confirm `gofmt` + `go vet` are clean, and that the change builds across the
   six 64-bit Go targets (amd64, arm64, riscv64, loong64, ppc64le, s390x) — the
   CI matrix does this on every push.
5. Open a PR in English, referencing the issue.

## Where things live

The parser — `Parse`, `Forward`, `Backward`, `ResolveSource`, `WithProjectRoot`,
and the `Record` type — is in
[`github.com/go-synctex/synctex`](https://github.com/go-synctex/synctex). This
documentation site is in
[`github.com/go-synctex/docs`](https://github.com/go-synctex/docs). Start from
the [Usage & API](api.md) page and the [Roadmap](roadmap.md) to find the right
place for your change.
