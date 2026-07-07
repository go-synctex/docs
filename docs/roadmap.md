# Roadmap

`go-synctex/synctex` is grown **test-first**, each capability covered before it
ships. The parser — the source↔PDF query core an editor or viewer needs — is
**complete**; the items below the line are honest "could come next", not
promises.

| Stage | What | Status |
| --- | --- | --- |
| Stream parser | Line-oriented scan of the SyncTeX stream: the `Input:` file-tag map, the `Content:` marker, `{P` / `}` page blocks, `Postamble:`, and the `[`, `h`, `v`, `k`, `g`, `x` record lines with their `N,L:X,Y` prefix. Malformed lines skipped, not fatal. | **Done** |
| Gzip + plain input | Reads `.synctex.gz` (gzip) and plain `.synctex`, selected on the file suffix. | **Done** |
| Forward query | `(source file, line)` → nearest `Record` (page + x/y in sp), binary-searched over a per-source line index built at parse time. | **Done** |
| Backward query | `(page, x, y in sp)` → `(source file, line)` of the nearest record on the page (smallest squared distance). | **Done** |
| Source resolution | `ResolveSource` maps an editor-side project-relative path to the absolute path SyncTeX recorded (exact / suffix / base-name match). | **Done** |
| Project-root sanitisation | `WithProjectRoot` drops `Input:` paths outside the root and rewrites inside paths relative-to-root, resolving symlinks first — so host paths never leak to a client. | **Done** |
| Zip-bomb hardening | Decompressed body read through a `LimitReader` and rejected past a 32 MB cap before scanning. | **Done** |
| Tests & coverage | 100% coverage including every error and malformed-record branch, `gofmt` + `go vet` clean, green across all six 64-bit Go arches (amd64, arm64, riscv64, loong64, ppc64le, s390x). | **Done** |

## Possible future work

These are **candidates**, not commitments — recorded so the surface's current
boundaries are explicit:

- **Typed sp / point helpers.** A small `ScaledPoint` type (or `ToPDFPoints`
  helper) so callers don't hand-divide by `65536`. Today the raw sp are exposed
  deliberately and the conversion is left to the caller; a typed convenience
  layer could sit on top without changing that contract.
- **Box-dimension parsing (`W,H,D`).** The parser reads the `N,L:X,Y` anchor and
  stops; the trailing width/height/depth are skipped because the current queries
  don't need them. Parsing them would enable box-extent-aware hit testing (e.g.
  "which box *contains* this click" rather than "which anchor is nearest").
- **Streaming API.** `Parse` reads the (capped) body into memory before scanning.
  A streaming variant that scans an `io.Reader` incrementally could lower the
  peak footprint for callers that already control the source.

## Documented out-of-scope boundaries

- **Coordinates are reported in raw scaled points.** The parser does not apply
  `Magnification:` / `Unit:` / `X Offset:` / `Y Offset:` transforms; callers that
  need them apply them at the boundary. See
  [The `.synctex.gz` format](format.md#the-preamble).
- **The `W,H,D` box dimensions are not currently parsed** — only the anchor
  `X,Y`. See the future-work note above.

See [Usage & API](api.md) for the surface and
[The `.synctex.gz` format](format.md) for the stream these stages build on.
