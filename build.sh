#!/usr/bin/env bash
set -e

DOC_MD="doc/document.md"
HTML_TPL="doc/html.txt"
AWK_SCRIPT="doc/go.awk"
PKG_NAME="gofpdf"

usage() {
  echo "Uso: $0 {all|documentation|cov|check|readme|html|docgo|build|clean}"
  exit 1
}

all() { documentation; }

documentation() {
  html
  docgo
  readme
}

cov() {
  all
  go test -v -coverprofile=coverage
  go tool cover -html=coverage -o=coverage.html
}

check() {
  golint .
  go vet -all .
  gofmt -s -l .
  goreportcard-cli -v | grep -v cyclomatic
}

readme() {
  pandoc --read=markdown --write=gfm < "$DOC_MD" > README.md
}

html() {
  pandoc --read=markdown --write=html --template="$HTML_TPL" \
    --metadata pagetitle="GoFPDF Document Generator" < "$DOC_MD" > doc/index.html
}

docgo() {
  pandoc --read=markdown --write=plain "$DOC_MD" | \
    awk --assign=package_name="$PKG_NAME" --file="$AWK_SCRIPT" > doc.go
  gofmt -s -w doc.go
}

build() {
  go build -v
}

clean() {
  rm -f coverage.html coverage doc/index.html doc.go README.md
}

# entrypoint
[ $# -ge 1 ] || usage
case "$1" in
  all|documentation|cov|check|readme|html|docgo|build|clean) "$1" ;;
  *) usage ;;
esac