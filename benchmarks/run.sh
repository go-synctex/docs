#!/usr/bin/env bash
#
# Copyright (c) the go-synctex authors
# SPDX-License-Identifier: BSD-3-Clause
#
# Library-level benchmark runner for go-synctex/synctex.
#
# Builds a synthetic .synctex.gz in memory (PAGES x PER_PAGE records) and times
# the pure-Go library's three primitives through its Go API — parse throughput,
# Forward latency, Backward latency — then prints a Markdown table of ns/op.
# There is no external oracle or reference runtime: the library performs no
# interpreter dispatch, so the meaningful numbers are its own operation costs.
#
# Usage:  bash benchmarks/run.sh
# Env:    OUTER (timed passes, default 25), WARM (untimed passes, default 3),
#         PAGES / PER_PAGE (synthetic corpus size).
set -u
cd "$(dirname "$0")"

command -v go >/dev/null 2>&1 || { echo "go not found on PATH" >&2; exit 1; }

TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

echo "== go-synctex/synctex library-level benchmark ==" >&2
echo "  go ..." >&2
( cd go && GOWORK=off go run . 2>/dev/null ) \
  | awk '$1=="RESULT"{printf "%s\t%s\n", $2, $3}' >> "$TMP"

echo >&2
awk -F'\t' '
  BEGIN {
    print  "| Operation | ns/op |"
    print  "| --- | ---: |"
    order = "parse forward backward"
    n = split(order, ord, " ")
  }
  { ns[$1] = $2 }
  END {
    for (o = 1; o <= n; o++) {
      k = ord[o]
      if (ns[k] == "") continue
      name = k
      if (k == "parse")    name = "parse (.synctex.gz -> indexed File)"
      if (k == "forward")  name = "Forward (source -> PDF)"
      if (k == "backward") name = "Backward (PDF -> source)"
      printf "| %s | %s |\n", name, ns[k]
    }
  }
' "$TMP"
