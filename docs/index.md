# go-synctex documentation

**A pure-Go (no cgo) parser for TeX's SyncTeX output** — the `.synctex.gz` file
`pdflatex` / `xelatex` / `lualatex` emit to link a `.tex` **source line** to the
exact **spot on the PDF page**, in both directions.

`go-synctex/synctex` reads a `.synctex.gz` (or plain `.synctex`) file and answers
the two queries an editor or PDF viewer needs — forward (source → PDF) and
backward (PDF → source) — with **zero external dependencies**, standard library
only. The module path is `github.com/go-synctex/synctex`.

The module is **standalone and reusable**: any Go program — a LaTeX editor, a PDF
preview server, a build pipeline — can import it. It cross-compiles and embeds
anywhere, and is hardened against hostile input.

!!! success "Status: parser complete — forward + backward"
    A line-oriented parser for the SyncTeX stream: the `Input:` file-tag map, the
    `Content:` page blocks, and the record lines (`[`, `h`, `v`, `k`, `g`, `x`)
    that pin a source line to a PDF coordinate. It answers **`Forward`**
    (source → PDF, binary-searched on the nearest line) and **`Backward`**
    (PDF → source, nearest point on a page), resolves editor-side source paths,
    and can scope untrusted `Input:` paths to a project root. Hardened with a
    32 MB decompression cap against zip-bomb `.synctex.gz` files. **100% test
    coverage** (including every error and malformed-record branch), `gofmt` +
    `go vet` clean, and green across the six 64-bit Go targets — amd64, arm64,
    riscv64, loong64, ppc64le, s390x.

## What it is

SyncTeX is the mechanism modern TeX engines use to record, as they typeset, the
correspondence between each box on the page and the source line that produced it.
The engine writes this correspondence to a `main.synctex.gz` next to the PDF.
`go-synctex/synctex` parses that file and turns it into two lookups:

- **Forward** — `(source file, line)` → `(page, x, y)`: jump the PDF viewer to
  where a source line was typeset.
- **Backward** — `(page, x, y)` → `(source file, line)`: click a glyph in the
  PDF and jump back to the source that produced it.

Coordinates are SyncTeX **scaled points** (sp); PDF points are `sp / 65536`. The
raw sp are exposed so callers apply their own DPI / point scaling.

## Quick taste

```go
package main

import (
	"fmt"

	"github.com/go-synctex/synctex"
)

func main() {
	// Parse the .synctex.gz the TeX engine wrote next to the PDF.
	f, err := synctex.Parse("main.synctex.gz")
	if err != nil {
		panic(err)
	}

	// Forward: source → PDF. "main.tex" line 42 lands here:
	if r, ok := f.Forward("main.tex", 42); ok {
		fmt.Printf("page %d at (%d, %d) sp\n", r.Page, r.X, r.Y)
	}

	// Backward: PDF → source. A click at (x, y) sp on page 2:
	if src, line, ok := f.Backward(2, 32_000_000, 45_000_000); ok {
		fmt.Printf("%s:%d\n", src, line)
	}
}
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`synctex`](https://github.com/go-synctex/synctex) | the parser — `Parse`, `Forward`, `Backward`, `ResolveSource`, `WithProjectRoot`, and the `Record` type |
| [`docs`](https://github.com/go-synctex/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-synctex.github.io`](https://github.com/go-synctex/go-synctex.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-synctex/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain, embeddable in any editor or viewer.
- **Zero external dependencies** — standard library only; nothing to vendor.
- **Hardened by default.** A decompression-size cap guards against zip-bomb
  `.synctex.gz` files, and `WithProjectRoot` keeps absolute host paths from
  leaking back to a client.
- **100% test coverage** is the target, enforced as a CI gate — every error and
  malformed-record branch included.

## Where to go next

- [Why a pure-Go SyncTeX parser](why.md) — CGO-free portability, the security
  angle, and the scaled-point coordinate model.
- [Usage & API](api.md) — `Parse`, `WithProjectRoot`, `Forward`, `Backward`,
  `ResolveSource`, `Record`, and the sp ↔ PDF-point conversion.
- [The `.synctex.gz` format](format.md) — what the file the parser consumes
  actually contains.
- [Roadmap](roadmap.md) — what is done and what could come next.

Source lives at
[github.com/go-synctex/synctex](https://github.com/go-synctex/synctex).
