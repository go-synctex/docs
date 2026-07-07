# Why a pure-Go SyncTeX parser

Linking a `.tex` source line to a spot on a PDF page — and back — is the core
feature every LaTeX editor and PDF previewer wants. TeX engines already compute
that correspondence and write it to a `.synctex.gz` file; the only missing piece
is a parser that turns the file into the two lookups the UI needs. The C
reference implementation (`synctex_parser.c`, shipped with TeX Live) does this,
but it is C — which means cgo, a build toolchain, and a shared library to link
and ship.

`go-synctex/synctex` is that parser reimplemented in **pure Go**, with **no cgo
and zero external dependencies**. It is a small, deterministic, standard-library
module you can drop into any Go program.

## CGO-free portability

Because the parser is CGO-free and depends only on the standard library, it:

- **cross-compiles to every Go target with no C toolchain** and links into a
  single static binary — no `synctex_parser.so`, no `libz` to find, no build
  matrix around a C dependency;
- **embeds anywhere** a Go program runs: a desktop editor, a PDF preview server,
  a CI step that validates SyncTeX output, a WASM build;
- **builds and passes its tests identically on every platform** — the same suite
  runs green across the six 64-bit Go targets (amd64, arm64, riscv64, loong64,
  ppc64le, s390x), with nothing platform-specific to special-case.

 Writing it in Go is not a rewrite for its own sake: the SyncTeX stream is a
line-oriented text format, and a tight Go scanner over it is both simpler and
safer than binding a C library through cgo.

## Embeddable in editors and viewers

The API is two functions and a struct: parse the file once, then call `Forward`
to drive a PDF viewer to a source line, or `Backward` to jump from a click on the
page back to the source. `Forward` binary-searches an index built at parse time,
so repeated queries — every keystroke that moves the cursor — are cheap. There is
no daemon, no IPC, no C process to manage: it is a library call in your own
address space.

`ResolveSource` bridges the editor's project-relative view of a file
(`main.tex`) to the absolute paths SyncTeX records under the compile workdir, so
the editor never has to know where the build actually ran.

## Hardened by default

A `.synctex.gz` is often produced by an **untrusted** build — a user-submitted
document, a shared compile server, a CI job. The parser treats it that way:

- **Zip-bomb cap.** The gzip body is read through a `LimitReader` and rejected if
  it exceeds **32 MB** decompressed, before it is scanned. A hostile or runaway
  `.synctex.gz` cannot inflate to exhaust memory. Real-world synctex files are
  sub-megabyte, so the cap is pure headroom.
- **Project-root scoping.** `Input:` entries in a SyncTeX file are absolute paths
  under the compile workdir. Echoed back to a client verbatim, they leak host
  filesystem layout. Pass
  [`WithProjectRoot`](api.md#withprojectroot) and paths that resolve outside the
  root are **dropped** (no record kept) while paths inside are **rewritten
  relative-to-root** — symlinks resolved before the comparison — so `Backward`
  and `ResolveSource` never hand an absolute host path back to the SPA.

The `WithProjectRoot` option is opt-in: the zero-value default preserves the
absolute paths for local, trusted callers, and untrusted-input handlers turn on
scoping explicitly.

## The scaled-point coordinate model

SyncTeX coordinates are **scaled points** (sp), TeX's internal length unit:
`1 pt = 65536 sp`. The parser preserves the raw sp in every `Record` rather than
pre-dividing to PDF points, because different consumers want different units:

- a PDF renderer such as **PDF.js** wants **PDF points** (`sp / 65536`), then
  scales by its own DPI / zoom;
- a coordinate-space transform may want the exact integer sp to avoid rounding
  drift across a chain of conversions.

Exposing sp keeps the library policy-free: it reports what SyncTeX recorded, and
the caller decides how to map it onto a rendered page. The conversion is a single
division — see [Usage & API](api.md#coordinates-scaled-points-vs-pdf-points).

See [Usage & API](api.md) for the surface and
[The `.synctex.gz` format](format.md) for the stream the parser reads.
