# Usage & API

The public API lives at the module root (`github.com/go-synctex/synctex`). It is
small and Go-idiomatic: parse a file once into a `*File`, then query it. Every
query returns an explicit `ok bool` (or `error`), value types, no global state.

!!! success "Status: implemented"
    `Parse`, `WithProjectRoot`, `Forward`, `Backward`, and `ResolveSource` are
    built and importable as `github.com/go-synctex/synctex`. See
    [Roadmap](roadmap.md) for what could come next.

## Install

```sh
go get github.com/go-synctex/synctex
```

## Parsing a file

```go
package main

import (
	"fmt"

	"github.com/go-synctex/synctex"
)

func main() {
	// Parse the .synctex.gz the TeX engine wrote next to the PDF.
	// A plain (uncompressed) .synctex file works too — the .gz suffix
	// selects gzip decompression.
	f, err := synctex.Parse("main.synctex.gz")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d records across %d source files\n",
		len(f.Records), len(f.Inputs))
}
```

`Parse` opens the file, decompresses it if the path ends in `.gz` (capped at
32 MB decompressed against zip-bomb inputs), scans the stream, and returns a
`*File` with the records indexed by source path so `Forward` can binary-search.

## Forward: source → PDF

```go
// "main.tex" line 42 → the nearest typeset point.
if r, ok := f.Forward("main.tex", 42); ok {
	fmt.Printf("page %d at (%d, %d) sp\n", r.Page, r.X, r.Y)
}
```

`Forward` resolves the source path, then binary-searches the file's records for
the **first record whose line ≥ the query**. If every record on the file is
before the query line, it returns the last one — so a query always lands
somewhere sensible as long as the file has any records.

## Backward: PDF → source

```go
// A click at (x, y) sp on page 2 → the closest source line.
if src, line, ok := f.Backward(2, 32_000_000, 45_000_000); ok {
	fmt.Printf("%s:%d\n", src, line)
}
```

`Backward` scans the records on the given page and returns the source file and
line of the record whose `(X, Y)` is **nearest** (smallest squared Euclidean
distance) to the queried point. The `x`, `y` arguments are in scaled points.

## ResolveSource: editor path → recorded path

```go
if abs, ok := f.ResolveSource("main.tex"); ok {
	fmt.Println(abs) // e.g. /srv/build/thesis/main.tex
}
```

`ResolveSource` maps an editor-side path to the path SyncTeX actually recorded in
its `Input:` entries. The editor typically has a project-relative name
(`main.tex`) while SyncTeX stores absolute paths under the compile workdir;
`ResolveSource` matches by exact path, by `/`-suffix, or by base name.
`Forward` calls it internally, so you rarely call it directly — it is exported
for callers that need the recorded path itself.

## WithProjectRoot

```go
// Scope untrusted Input: paths to a project root.
f, err := synctex.Parse("build/main.synctex.gz",
	synctex.WithProjectRoot("/srv/projects/thesis"))
```

`WithProjectRoot` scopes the `Input:` paths to the given root:

- paths that resolve **outside** the root are **dropped** — no record is kept for
  them, so they can never be echoed back to a client;
- paths **inside** the root are **rewritten relative-to-root**, so `Backward` and
  `ResolveSource` return project-relative paths instead of absolute host paths.

Symlinks on both the root and the candidate are resolved (via `EvalSymlinks`)
before the containment check, so a symlink pointing out of the root
(`.workspace/leak → /etc`) is rejected. Synthetic paths the engine emits but
never materialises on disk fall back to a lexical clean. When the option is
**not** passed (the zero value), paths are `filepath.Clean`'d but otherwise left
as absolute — the right default for local, trusted callers.

## Coordinates: scaled points vs PDF points

SyncTeX coordinates are **scaled points** (sp): `1 pt = 65536 sp`. `Record.X` and
`Record.Y` hold the raw sp; convert to PDF points at the boundary:

```go
r, _ := f.Forward("main.tex", 42)
pdfX := float64(r.X) / 65536.0 // PDF points
pdfY := float64(r.Y) / 65536.0
// A renderer (e.g. PDF.js) then scales pdfX/pdfY by its own DPI / zoom.
```

The library deliberately keeps the raw sp so callers can choose their own unit
and avoid rounding drift — see
[Why a pure-Go SyncTeX parser](why.md#the-scaled-point-coordinate-model).

## Shape

### `Parse`

```go
// Parse reads a .synctex.gz (or plain .synctex) file from disk and returns a
// File with its records indexed for Forward queries.
func Parse(path string, opts ...Option) (*File, error)
```

- **`path`** — the `.synctex.gz` (gzip) or `.synctex` (plain) file. The `.gz`
  suffix selects decompression.
- **`opts`** — zero or more `Option`s; currently `WithProjectRoot`.
- **`error`** — non-nil for an unreadable file, a bad gzip stream, or a body that
  exceeds the 32 MB decompression cap.

### `WithProjectRoot` and `Option`

```go
// Option tunes Parse behaviour.
type Option func(*parseOptions)

// WithProjectRoot scopes Input: paths to root: outside paths are dropped,
// inside paths are rewritten relative-to-root.
func WithProjectRoot(root string) Option
```

### `File`

```go
type File struct {
	Inputs  map[int]string // Input: tag → source path (relative-to-root under WithProjectRoot)
	Records []Record       // every parsed hit point, in stream order
	// (an internal per-source line index powers Forward's binary search)
}

// Forward: (source file, line) → nearest Record (page + x/y in sp).
func (f *File) Forward(sourceFile string, line int) (Record, bool)

// Backward: (page, x, y in sp) → (source file, line) closest to the point.
func (f *File) Backward(page, x, y int) (sourceFile string, line int, ok bool)

// ResolveSource maps an editor-side path to the path SyncTeX recorded.
func (f *File) ResolveSource(rel string) (string, bool)
```

### `Record`

```go
type Record struct {
	FileTag int // Input: tag this record's source maps through
	Line    int // 1-based source line
	Page    int // 1-based PDF page
	X, Y    int // anchor, in SyncTeX scaled points (sp); PDF points = sp / 65536
}
```

`Inputs` maps each `FileTag` to its source path; `Forward` / `Backward` do that
lookup for you and hand back the path directly.
