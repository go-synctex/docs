<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-synctex` library-level benchmark harness

Reproducible, self-contained benchmark of the **pure-Go `go-synctex/synctex`
library**. It measures the library's three primitives through its Go API —
**parse throughput**, **`Forward` latency**, **`Backward` latency** — so the
numbers answer: *how fast does the parser turn a `.synctex.gz` into an indexed
file, and how cheap are the source↔PDF queries afterwards?*

There is **no external oracle or reference runtime**: the library performs no
interpreter dispatch and has nothing to compare against. The meaningful figures
are its own operation costs.

## What is measured

- **parse** — `synctex.Parse` on a synthetic `.synctex.gz`: decompress (capped at
  32 MB), scan the stream, build the per-source line index. The one-time cost
  paid when a document is (re)built.
- **forward** — `File.Forward(source, line)`: resolve the source path, then a
  binary search over the sorted records for that file. The interactive path an
  editor hits on every cursor move.
- **backward** — `File.Backward(page, x, y)`: a nearest-anchor scan of the records
  on the queried page. The path a PDF click takes.

## Synthetic corpus

The harness builds its input **in memory** — no fixture files, no TeX
installation. It renders a deterministic SyncTeX stream of `PAGES` page blocks,
each holding `PER_PAGE` record lines with generated file tags, source lines, and
scaled-point coordinates, then gzip-compresses it exactly as the engine would.
Because it is generated deterministically, a run on any of the six supported
arches measures the same workload.

## Layout

- `go/`     — the self-contained Go driver. `main.go` builds the synthetic
  corpus and defines the three benchmarked ops; `bench.go` is the shared timer.
  `go.mod` pins the published library by version (no `replace`). Any built
  `go/bench` binary is git-ignored.
- `run.sh`  — runs the Go driver and prints a Markdown table of ns/op.

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `PAGES` / `PER_PAGE` to scale the synthetic corpus.

## Verify

The driver accepts a `verify` argument that prints a canonical summary of the
synthetic corpus (record count, a `Forward` result, a `Backward` result), so a
run can be sanity-checked before any timing is trusted:

```sh
(cd benchmarks/go && GOWORK=off go run . verify)
```

## Method

Each run builds the synthetic `.synctex.gz` once, materialises it to a temp file
(so `Parse` reads from disk as a real caller would), then for each op runs `WARM`
untimed passes followed by `OUTER` timed passes of a fixed inner loop, timed with
a monotonic clock; the **best** pass is reported as **ns/op**. The query
benchmarks (`forward`, `backward`) run against a pre-parsed `*File` so their
timing excludes the parse. Results characterise the corpus size you dial in on
the host you run on — re-run the harness rather than quoting a figure; see
`../docs/performance.md`.
