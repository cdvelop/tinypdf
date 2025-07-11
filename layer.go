package tinypdf

// Routines in this file are translated from
// http://www.tinypdf.org/en/script/script97.php

type layerType struct {
	name    string
	visible bool
	objNum  int // object number
}

type layerRecType struct {
	list          []layerType
	currentLayer  int
	openLayerPane bool
}

func (f *DocPDF) layerInit() {
	f.layer.list = make([]layerType, 0)
	f.layer.currentLayer = -1
	f.layer.openLayerPane = false
}

// AddLayer defines a layer that can be shown or hidden when the document is
// displayed. name specifies the layer name that the document reader will
// display in the layer list. visible specifies whether the layer will be
// initially visible. The return value is an integer ID that is used in a call
// to BeginLayer().
func (f *DocPDF) AddLayer(name string, visible bool) (layerID int) {
	layerID = len(f.layer.list)
	f.layer.list = append(f.layer.list, layerType{name: name, visible: visible})
	return
}

// BeginLayer is called to begin adding content to the specified layer. All
// content added to the page between a call to BeginLayer and a call to
// EndLayer is added to the layer specified by id. See AddLayer for more
// details.
func (f *DocPDF) BeginLayer(id int) {
	f.EndLayer()
	if id >= 0 && id < len(f.layer.list) {
		f.outf("/OC /OC%d BDC", id)
		f.layer.currentLayer = id
	}
}

// EndLayer is called to stop adding content to the currently active layer. See
// BeginLayer for more details.
func (f *DocPDF) EndLayer() {
	if f.layer.currentLayer >= 0 {
		f.out("EMC")
		f.layer.currentLayer = -1
	}
}

// OpenLayerPane advises the document reader to open the layer pane when the
// document is initially displayed.
func (f *DocPDF) OpenLayerPane() {
	f.layer.openLayerPane = true
}

func (f *DocPDF) layerEndDoc() {
	if len(f.layer.list) == 0 {
		return
	}
	if f.pdfVersion < pdfVers1_5 {
		f.pdfVersion = pdfVers1_5
	}
}

func (f *DocPDF) layerPutLayers() {
	for j, l := range f.layer.list {
		f.newobj()
		f.layer.list[j].objNum = f.n
		f.outf("<</Type /OCG /Name %s>>", f.textstring(utf8toutf16(l.name)))
		f.out("endobj")
	}
}

func (f *DocPDF) layerPutResourceDict() {
	if len(f.layer.list) > 0 {
		f.out("/Properties <<")
		for j, layer := range f.layer.list {
			f.outf("/OC%d %d 0 R", j, layer.objNum)
		}
		f.out(">>")
	}

}

func (f *DocPDF) layerPutCatalog() {
	if len(f.layer.list) > 0 {
		onStr := ""
		offStr := ""
		for _, layer := range f.layer.list {
			onStr += sprintf("%d 0 R ", layer.objNum)
			if !layer.visible {
				offStr += sprintf("%d 0 R ", layer.objNum)
			}
		}
		f.outf("/OCProperties <</OCGs [%s] /D <</OFF [%s] /Order [%s]>>>>", onStr, offStr, onStr)
		if f.layer.openLayerPane {
			f.out("/PageMode /UseOC")
		}
	}
}
