# Performance

`go-synctex/synctex` is a **parse-once, query-many** library. The cost model is
simple and worth understanding because it tells you where to spend and where you
get things for free:

- **Parse** — a one-time cost per file: decompress (capped at 32 MB), scan the
  stream line by line into `Records`, then build a per-source line index sorted
  by line. This dominates the total, and it happens once when the PDF is (re)built.
- **Forward** — cheap and repeatable: resolve the source path, then a
  **binary search** (`sort.Search`) over that file's sorted records — `O(log n)`
  in the number of records on the file. This is the hot path an editor hits on
  every cursor move, and it is designed to be effectively free after the parse.
- **Backward** — a **linear scan** of the records on the queried page,
  `O(records on page)`, picking the nearest anchor by squared distance. It runs
  on a user click, not per keystroke, so a scan is fine; there is no per-query
  allocation.

The design choice is deliberate: pay the indexing cost **once** at parse time so
the interactive `Forward` query is a logarithmic search rather than a scan.

## What to measure

The library performs no interpreter dispatch and has no external oracle, so there
is nothing to compare it *against* — the meaningful numbers are its own
throughput and latency:

- **Parse throughput** — MB/s (or records/s) turning a `.synctex.gz` into an
  indexed `*File`, the figure that matters for large documents and rebuild loops.
- **`Forward` latency** — ns/op for a source→PDF lookup on an already-parsed
  file; the interactive-path number.
- **`Backward` latency** — ns/op for a PDF→source lookup, which scales with the
  record count on the target page.

## Benchmark harness

A self-contained Go harness lives under
[`benchmarks/`](https://github.com/go-synctex/docs/tree/main/benchmarks). It
**generates a synthetic large `.synctex.gz` in memory** — a configurable number
of pages and records per page, gzip-compressed exactly as the TeX engine would
write it — then times:

- `synctex.Parse` on the synthetic file (parse + index throughput),
- `File.Forward` for a fixed `(source, line)` query (interactive latency),
- `File.Backward` for a fixed `(page, x, y)` query (click latency).

Run it with:

```sh
bash benchmarks/run.sh
```

Environment knobs `PAGES` and `PER_PAGE` scale the synthetic corpus, and
`OUTER` / `WARM` tune the timed / warm-up pass budget. See
[`benchmarks/README.md`](https://github.com/go-synctex/docs/tree/main/benchmarks)
for the full harness description.

!!! note "Reproducible, self-contained"
    The harness builds its own input — no fixture files, no external TeX
    installation, no reference runtime. The synthetic `.synctex.gz` is generated
    deterministically in memory, so a run on any of the six supported arches
    measures the same workload. `go.mod` pins the published library by version
    (no `replace`), so the benchmark exercises exactly what a consumer imports.

!!! warning "Honest framing"
    The harness measures **operation cost** — parse, `Forward`, `Backward` — on a
    synthetic corpus, not a specific real document, so absolute ns/op depend on
    the corpus size you dial in and the host you run on. The shape is what
    matters: parse is the one-time cost, `Forward` is `O(log n)`, `Backward` is
    `O(records on page)`. Treat the numbers as characterising *your* corpus on
    *your* hardware, and re-run the harness rather than quoting a figure from
    memory.
