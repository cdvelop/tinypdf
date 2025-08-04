package tinypdf

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"

	. "github.com/cdvelop/tinystring"
)

// newTpl creates a template, copying graphics settings from a template if one is given
func newTpl(corner PointType, size SizeType, orientationStr orientationType, unitType unit, fontsDirName FontsDirName, fn func(*Tpl), copyFrom *DocPDF) Template {
	pageSize := PageSize{Wd: size.Wd, Ht: size.Ht, AutoHt: false}

	docpdf := New(unitType, orientationStr, pageSize, fontsDirName)
	tpl := Tpl{*docpdf}
	if copyFrom != nil {
		tpl.loadParamsFromFpdf(copyFrom)
	}
	tpl.DocPDF.AddPage()
	fn(&tpl)

	bytes := make([][]byte, len(tpl.DocPDF.pages))
	// skip the first page as it will always be empty
	for x := 1; x < len(bytes); x++ {
		bytes[x] = tpl.DocPDF.pages[x].Bytes()
	}

	templates := make([]Template, 0, len(tpl.DocPDF.templates))
	for _, key := range templateKeyList(tpl.DocPDF.templates, true) {
		templates = append(templates, tpl.DocPDF.templates[key])
	}
	images := tpl.DocPDF.images

	template := FpdfTpl{corner, size, bytes, images, templates, tpl.DocPDF.page}
	return &template
}

// FpdfTpl is a concrete implementation of the Template interface.
type FpdfTpl struct {
	corner    PointType
	size      SizeType
	bytes     [][]byte
	images    map[string]*ImageInfoType
	templates []Template
	page      int
}

// ID returns the global template identifier
func (t *FpdfTpl) ID() string {
	return Fmt("%x", sha1.Sum(t.Bytes()))
}

// Size gives the bounding dimensions of this template
func (t *FpdfTpl) Size() (corner PointType, size SizeType) {
	return t.corner, t.size
}

// Bytes returns the actual template data, not including resources
func (t *FpdfTpl) Bytes() []byte {
	return t.bytes[t.page]
}

// FromPage creates a new template from a specific Page
func (t *FpdfTpl) FromPage(page int) (Template, error) {
	// pages start at 1
	if page == 0 {
		return nil, Err(D.Invalid, "docpdf: pages start at 1. No template will have a page 0")
	}

	if page > t.NumPages() {
		return nil, Err(D.Invalid, Fmt("docpdf: the template does not have a page %d", page))
	}
	// if it is already pointing to the correct page
	// there is no need to create a new template
	if t.page == page {
		return t, nil
	}

	t2 := *t
	t2.page = page
	return &t2, nil
}

// FromPages creates a template slice with all the pages within a template.
func (t *FpdfTpl) FromPages() []Template {
	p := make([]Template, t.NumPages())
	for x := 1; x <= t.NumPages(); x++ {
		// the only error is when accessing a
		// non existing template... that can't happen
		// here
		p[x-1], _ = t.FromPage(x)
	}

	return p
}

// Images returns a list of the images used in this template
func (t *FpdfTpl) Images() map[string]*ImageInfoType {
	return t.images
}

// Templates returns a list of templates used in this template
func (t *FpdfTpl) Templates() []Template {
	return t.templates
}

// NumPages returns the number of available pages within the template. Look at FromPage and FromPages on access to that content.
func (t *FpdfTpl) NumPages() int {
	// the first page is empty to
	// make the pages begin at one
	return len(t.bytes) - 1
}

// Serialize turns a template into a byte string for later deserialization
func (t *FpdfTpl) Serialize() ([]byte, error) {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(t)

	return b.Bytes(), err
}

// DeserializeTemplate creaties a template from a previously serialized
// template
func DeserializeTemplate(b []byte) (Template, error) {
	tpl := new(FpdfTpl)
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	err := dec.Decode(tpl)
	return tpl, err
}

// childrenImages returns the next layer of children images, it doesn't dig into
// children of children. Applies template namespace to keys to ensure
// no collisions. See UseTemplateScaled
func (t *FpdfTpl) childrenImages() map[string]*ImageInfoType {
	childrenImgs := make(map[string]*ImageInfoType)

	for x := 0; x < len(t.templates); x++ {
		imgs := t.templates[x].Images()
		for key, val := range imgs {
			name := Fmt("t%s-%s", t.templates[x].ID(), key)
			childrenImgs[name] = val
		}
	}

	return childrenImgs
}

// childrensTemplates returns the next layer of children templates, it doesn't dig into
// children of children.
func (t *FpdfTpl) childrensTemplates() []Template {
	childrenTmpls := make([]Template, 0)

	for x := 0; x < len(t.templates); x++ {
		tmpls := t.templates[x].Templates()
		childrenTmpls = append(childrenTmpls, tmpls...)
	}

	return childrenTmpls
}

// GobEncode encodes the receiving template into a byte buffer. Use GobDecode
// to decode the byte buffer back to a template.
func (t *FpdfTpl) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	childrensTemplates := t.childrensTemplates()
	firstClassTemplates := make([]Template, 0)

found_continue:
	for x := 0; x < len(t.templates); x++ {
		for y := 0; y < len(childrensTemplates); y++ {
			if childrensTemplates[y].ID() == t.templates[x].ID() {
				continue found_continue
			}
		}

		firstClassTemplates = append(firstClassTemplates, t.templates[x])
	}
	err := encoder.Encode(firstClassTemplates)

	childrenImgs := t.childrenImages()
	firstClassImgs := make(map[string]*ImageInfoType)

	for key, img := range t.images {
		if _, ok := childrenImgs[key]; !ok {
			firstClassImgs[key] = img
		}
	}

	if err == nil {
		err = encoder.Encode(firstClassImgs)
	}
	if err == nil {
		err = encoder.Encode(t.corner)
	}
	if err == nil {
		err = encoder.Encode(t.size)
	}
	if err == nil {
		err = encoder.Encode(t.bytes)
	}
	if err == nil {
		err = encoder.Encode(t.page)
	}

	return w.Bytes(), err
}

// GobDecode decodes the specified byte buffer into the receiving template.
func (t *FpdfTpl) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	firstClassTemplates := make([]*FpdfTpl, 0)
	err := decoder.Decode(&firstClassTemplates)
	t.templates = make([]Template, len(firstClassTemplates))

	for x := 0; x < len(t.templates); x++ {
		t.templates[x] = Template(firstClassTemplates[x])
	}

	firstClassImages := t.childrenImages()

	t.templates = append(t.childrensTemplates(), t.templates...)

	t.images = make(map[string]*ImageInfoType)
	if err == nil {
		err = decoder.Decode(&t.images)
	}

	for k, v := range firstClassImages {
		t.images[k] = v
	}

	if err == nil {
		err = decoder.Decode(&t.corner)
	}
	if err == nil {
		err = decoder.Decode(&t.size)
	}
	if err == nil {
		err = decoder.Decode(&t.bytes)
	}
	if err == nil {
		err = decoder.Decode(&t.page)
	}

	return err
}

// Tpl is an DocPDF used for writing a template. It has most of the facilities of
// an DocPDF, but cannot add more pages. Tpl is used directly only during the
// limited time a template is writable.
type Tpl struct {
	DocPDF
}

func (t *Tpl) loadParamsFromFpdf(f *DocPDF) {
	t.DocPDF.compress = false

	t.DocPDF.k = f.k
	t.DocPDF.x = f.x
	t.DocPDF.y = f.y
	t.DocPDF.lineWidth = f.lineWidth
	t.DocPDF.capStyle = f.capStyle
	t.DocPDF.joinStyle = f.joinStyle

	t.DocPDF.color.draw = f.color.draw
	t.DocPDF.color.fill = f.color.fill
	t.DocPDF.color.text = f.color.text

	t.DocPDF.fonts = f.fonts
	t.DocPDF.currentFont = f.currentFont
	t.DocPDF.fontFamily = f.fontFamily
	t.DocPDF.fontSize = f.fontSize
	t.DocPDF.fontSizePt = f.fontSizePt
	t.DocPDF.fontStyle = f.fontStyle
	t.DocPDF.ws = f.ws

	for key, value := range f.images {
		t.DocPDF.images[key] = value
	}
}
