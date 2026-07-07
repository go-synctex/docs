# The `.synctex.gz` format

This page describes the SyncTeX stream `go-synctex/synctex` consumes — enough to
understand what the parser reads and what it ignores. The authoritative reference
is Jérôme Laurens' *SyncTeX* documentation and the `synctex_parser` sources
shipped with TeX Live[^ref]; this page is scoped to the subset the parser
actually acts on.

[^ref]: Jérôme Laurens, *SyncTeX: introduction to the engine*, and the reference
    `synctex_parser.c` / `synctex_parser.h` in TeX Live. See also
    <https://www.tug.org/TUGboat/tb29-3/tb93laurens.pdf>.

## What the engine writes

When you compile with SyncTeX enabled (`pdflatex -synctex=1 main.tex`, the
default in most editors), the TeX engine writes `main.synctex.gz` next to the
PDF. It is a **gzip-compressed, line-oriented text file** recording, box by box,
which source line produced which point on which page. A plain (uncompressed)
`main.synctex` is the same content without the gzip wrapper — `Parse` handles
both, selecting decompression on the `.gz` suffix.

## Overall shape

```
SyncTeX Version:1
Input:1:/srv/build/thesis/main.tex
Input:2:/srv/build/thesis/chapter1.tex
Magnification:1000
Unit:1
X Offset:0
Y Offset:0
Content:
{1
[1,23:4736286,42152922:...
h 1,23:4736286,42152922:...
x 1,25:4736286,40000000:...
}
{2
[2,7:4736286,42152922:...
}
Postamble:
...
```

## The preamble

Before `Content:`, the file carries header lines. The parser cares about exactly
two of them:

- **`Input:N:/path/to/file.tex`** — declares that **file tag `N`** refers to the
  given source path. Every record later refers to a source file by this integer
  tag; the parser builds the tag → path map from these lines. Under
  [`WithProjectRoot`](api.md#withprojectroot) a path outside the root is dropped
  and a path inside is rewritten relative-to-root.
- **`Content:`** — marks the end of the preamble; everything after it is the body
  of page blocks and records.

The remaining header lines — **`SyncTeX Version:`**, **`Magnification:`**,
**`Unit:`**, **`X Offset:`**, **`Y Offset:`** — are part of the format but are
**not** interpreted by this parser: it works directly in scaled points and does
not apply magnification or offset transforms. Callers that need those transforms
apply them at the boundary.

## The content body

After `Content:`, the stream is a sequence of **page blocks** holding **record
lines**:

- **`{P`** — opens the block for **page `P`** (1-based). Records that follow
  belong to this page until it closes.
- **`}`** — closes the current page block.
- **`Postamble:`** — ends the content; the parser stops here.

### Record lines

Inside a page block, each line begins with a single-character node type followed
by the node's tag, line, and coordinate. The parser recognises these leading
characters as line-pinning records:

| Leading | Node | Form |
| --- | --- | --- |
| `[` | box open (hbox/vbox open) | `[N,L:X,Y:W,H,D` |
| `h` | horizontal box | `h N,L:X,Y:W,H,D` |
| `v` | vertical box | `v N,L:X,Y:W,H,D` |
| `k` | kern | `k N,L:X,Y:...` |
| `g` | glue | `g N,L:X,Y:...` |
| `x` | rule | `x N,L:X,Y:...` |

The `[` node's fields follow immediately; the `h`, `v`, `k`, `g`, `x` nodes carry
a single space before their fields (`h 1,23:...`). Any other leading character
(for example the box-close `]`, or void nodes the parser doesn't model) is
skipped.

### The `N,L:X,Y` prefix

Every record's payload starts with the same prefix, and that prefix is all the
parser needs:

```
N , L : X , Y : W , H , D
│   │   │   │   └── further box dimensions (width, height, depth) — ignored
│   │   │   └────── Y coordinate, in scaled points (sp)
│   │   └────────── X coordinate, in scaled points (sp)
│   └────────────── source line (1-based)
└────────────────── file tag N — indexes into the Input: map
```

The parser reads `N`, `L`, `X`, and `Y`, stopping at the second `:` (or a comma):
the trailing **`W,H,D`** box dimensions — width, height, depth — are **not**
parsed, because the forward/backward queries only need the anchor coordinate, not
the box extent. Each successfully parsed prefix becomes a
[`Record`](api.md#record) `{FileTag: N, Line: L, Page: P, X, Y}`.

A malformed record line — a missing comma, a non-numeric field — is **skipped**,
not fatal: the parser keeps scanning so one bad line never aborts the whole file.

## What the parser produces

From this stream the parser builds:

- **`Inputs`** — the tag → source-path map from the `Input:` lines;
- **`Records`** — one `Record` per parsed line-pinning node, in stream order;
- an internal **per-source line index**, sorted by line, so
  [`Forward`](api.md#forward-source-pdf) can binary-search the nearest line.

See [Usage & API](api.md) for how to query the result.
