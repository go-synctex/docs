// SPDX-License-Identifier: BSD-3-Clause
package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-synctex/synctex"
)

// The synthetic corpus is a plausibly-large .synctex.gz built entirely in
// memory: PAGES page blocks, each holding PER_PAGE record lines whose file tag,
// source line, and (x, y) scaled-point coordinates are generated
// deterministically — no fixture files, no TeX installation. This mirrors the
// shape the engine writes so synctex.Parse does real work.
var (
	pages   = envInt("PAGES", 500)
	perPage = envInt("PER_PAGE", 200)
)

// sourcePath is the single Input: file the synthetic records map through.
const sourcePath = "/synthetic/thesis/main.tex"

// buildSyncTeX renders a deterministic SyncTeX stream and gzip-compresses it,
// returning the .synctex.gz bytes.
func buildSyncTeX() []byte {
	var raw bytes.Buffer
	raw.WriteString("SyncTeX Version:1\n")
	raw.WriteString("Input:1:" + sourcePath + "\n")
	raw.WriteString("Magnification:1000\nUnit:1\nX Offset:0\nY Offset:0\n")
	raw.WriteString("Content:\n")
	for p := 1; p <= pages; p++ {
		raw.WriteString("{" + strconv.Itoa(p) + "\n")
		for i := 0; i < perPage; i++ {
			line := (p-1)*perPage + i + 1
			x := 4736286 + i*3200
			y := 42152922 - i*1800
			// h N,L:X,Y:W,H,D — a horizontal box pinning source line -> point.
			raw.WriteString(fmt.Sprintf("h 1,%d:%d,%d:0,0,0\n", line, x, y))
		}
		raw.WriteString("}\n")
	}
	raw.WriteString("Postamble:\n")

	var gz bytes.Buffer
	w := gzip.NewWriter(&gz)
	if _, err := w.Write(raw.Bytes()); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return gz.Bytes()
}

// writeCorpus writes the synthetic .synctex.gz to a temp file and returns its
// path; Parse reads from disk, so the corpus is materialised once up front.
func writeCorpus() string {
	path := filepath.Join(os.TempDir(), "synctex-bench.synctex.gz")
	if err := os.WriteFile(path, buildSyncTeX(), 0o644); err != nil {
		panic(err)
	}
	return path
}

func mustParse(path string) *synctex.File {
	f, err := synctex.Parse(path)
	if err != nil {
		panic(err)
	}
	return f
}

// verify prints a small canonical summary so a run can be sanity-checked.
func verify(path string) {
	f := mustParse(path)
	fmt.Printf("records=%d inputs=%d\n", len(f.Records), len(f.Inputs))
	midLine := (pages * perPage) / 2
	if r, ok := f.Forward("main.tex", midLine); ok {
		fmt.Printf("forward line %d -> page %d (%d,%d) sp\n", midLine, r.Page, r.X, r.Y)
	}
	if src, line, ok := f.Backward(pages/2, 4736286, 42152922); ok {
		fmt.Printf("backward -> %s:%d\n", src, line)
	}
}

func main() {
	path := writeCorpus()
	defer os.Remove(path)

	if len(os.Args) > 1 && os.Args[1] == "verify" {
		verify(path)
		return
	}

	// Pre-parse a file for the query benchmarks so their timing excludes parse.
	f := mustParse(path)
	midLine := (pages * perPage) / 2
	midPage := pages / 2

	// parse: decompress + scan + index a full .synctex.gz (one-time cost).
	bench("parse", 50, func() { sink = mustParse(path) })
	// forward: source -> PDF, binary-searched on the pre-parsed file.
	bench("forward", 200000, func() {
		r, _ := f.Forward("main.tex", midLine)
		sink = r
	})
	// backward: PDF -> source, nearest anchor on the page.
	bench("backward", 20000, func() {
		src, line, _ := f.Backward(midPage, 4736286, 42152922)
		sink = src
		sink = line
	})
}
