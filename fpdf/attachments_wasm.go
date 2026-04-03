// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//go:build wasm

package fpdf

import (
	. "github.com/tinywasm/fmt"
)

// Attachment defines a content to be included in the pdf
type Attachment struct {
	Content      []byte
	Filename     string
	Description  string
	objectNumber int
}

func (f *Fpdf) writeCompressedFileObject(content []byte) {
	f.err = Err("attachments", "not supported in WASM")
}

func (f *Fpdf) embed(a *Attachment) {
	f.err = Err("attachments", "not supported in WASM")
}

func (f *Fpdf) SetAttachments(as []Attachment) {
	f.err = Err("attachments", "not supported in WASM")
}

func (f *Fpdf) putAttachments() {
	// Stub for WASM
}

func (f Fpdf) getEmbeddedFiles() string {
	return ""
}

type annotationAttach struct {
	*Attachment
	x, y, w, h float64
}

func (f *Fpdf) AddAttachmentAnnotation(a *Attachment, x, y, w, h float64) {
	f.err = Err("attachments", "not supported in WASM")
}

func (f *Fpdf) putAnnotationsAttachments() {
	// Stub for WASM
}

func (f *Fpdf) putAttachmentAnnotationLinks(out *fmtBuffer, page int) {
	// Stub for WASM
}
