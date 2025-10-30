package tinypdf

// GetJavascript returns the Adobe JavaScript for the document.
//
// GetJavascript returns an empty string if no javascript was
// previously defined.
func (f *Fpdf) GetJavascript() string {
	if f.javascript == nil {
		return ""
	}
	return *f.javascript
}

// SetJavascript adds Adobe JavaScript to the document.
func (f *Fpdf) SetJavascript(script string) {
	f.javascript = &script
}

func (f *Fpdf) putjavascript() {
	if f.javascript == nil {
		return
	}

	f.newobj()
	f.nJs = f.n
	f.out("<<")
	f.outf("/Names [(EmbeddedJS) %d 0 R]", f.n+1)
	f.out(">>")
	f.out("endobj")
	f.newobj()
	f.out("<<")
	f.out("/S /JavaScript")
	f.outf("/JS %s", f.textstring(*f.javascript))
	f.out(">>")
	f.out("endobj")
}
