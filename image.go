package tinypdf

import (
	"bufio"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

type imgInfo struct {
	w, h int
	//src              string
	formatName       string
	colspace         string
	bitsPerComponent string
	filter           string
	decodeParms      string
	trns             []byte
	smask            []byte
	smarkObjID       int
	pal              []byte
	deviceRGBObjID   int
	data             []byte
}

// ImageByHolder : draw image by ImageHolder
func (gp *GoPdf) ImageByHolder(img ImageHolder, x float64, y float64, rect *Rect) error {
	gp.UnitsToPointsVar(&x, &y)

	rect = rect.UnitsToPoints(gp.config.Unit)

	imageOptions := ImageOptions{
		X:    x,
		Y:    y,
		Rect: rect,
	}

	return gp.imageByHolder(img, imageOptions)
}

func (gp *GoPdf) ImageByHolderWithOptions(img ImageHolder, opts ImageOptions) error {
	gp.UnitsToPointsVar(&opts.X, &opts.Y)

	opts.Rect = opts.Rect.UnitsToPoints(gp.config.Unit)

	imageTransparency, err := gp.getCachedTransparency(opts.Transparency)
	if err != nil {
		return err
	}

	if imageTransparency != nil {
		opts.extGStateIndexes = append(opts.extGStateIndexes, imageTransparency.extGStateIndex)
	}

	if opts.Mask != nil {
		maskTransparency, err := gp.getCachedTransparency(opts.Mask.ImageOptions.Transparency)
		if err != nil {
			return err
		}

		if maskTransparency != nil {
			opts.Mask.ImageOptions.extGStateIndexes = append(opts.Mask.ImageOptions.extGStateIndexes, maskTransparency.extGStateIndex)
		}

		gp.UnitsToPointsVar(&opts.Mask.ImageOptions.X, &opts.Mask.ImageOptions.Y)
		opts.Mask.ImageOptions.Rect = opts.Mask.ImageOptions.Rect.UnitsToPoints(gp.config.Unit)

		extGStateIndex, err := gp.maskHolder(opts.Mask.Holder, *opts.Mask)
		if err != nil {
			return err
		}

		opts.extGStateIndexes = append(opts.extGStateIndexes, extGStateIndex)
	}

	return gp.imageByHolder(img, opts)
}

func (gp *GoPdf) maskHolder(img ImageHolder, opts MaskOptions) (int, error) {
	var cacheImage *ImageCache
	var cacheContentImage *cacheContentImage

	for _, imgcache := range gp.curr.ImgCaches {
		if img.ID() == imgcache.Path {
			cacheImage = &imgcache
			break
		}
	}

	if cacheImage == nil {
		maskImgobj := &ImageObj{IsMask: true}
		maskImgobj.init(func() *GoPdf {
			return gp
		})
		maskImgobj.setProtection(gp.protection())

		err := maskImgobj.SetImage(img)
		if err != nil {
			return 0, err
		}

		if opts.Rect == nil {
			if opts.Rect, err = maskImgobj.getRect(); err != nil {
				return 0, err
			}
		}

		if err := maskImgobj.parse(); err != nil {
			return 0, err
		}

		if gp.indexOfProcSet != -1 {
			index := gp.addObj(maskImgobj)
			cacheContentImage = gp.getContent().GetCacheContentImage(index, opts.ImageOptions)
			procset := gp.pdfObjs[gp.indexOfProcSet].(*ProcSetObj)
			procset.RelateXobjs = append(procset.RelateXobjs, RelateXobject{IndexOfObj: index})

			imgcache := ImageCache{
				Index: index,
				Path:  img.ID(),
				Rect:  opts.Rect,
			}
			gp.curr.ImgCaches[index] = imgcache
			gp.curr.CountOfImg++
		}
	} else {
		if opts.Rect == nil {
			opts.Rect = gp.curr.ImgCaches[cacheImage.Index].Rect
		}

		cacheContentImage = gp.getContent().GetCacheContentImage(cacheImage.Index, opts.ImageOptions)
	}

	if cacheContentImage != nil {
		extGStateInd, err := gp.createTransparencyXObjectGroup(cacheContentImage, opts)
		if err != nil {
			return 0, err
		}

		return extGStateInd, nil
	}

	return 0, errUndefinedCacheContentImage
}

func (gp *GoPdf) createTransparencyXObjectGroup(image *cacheContentImage, opts MaskOptions) (int, error) {
	bbox := opts.BBox
	if bbox == nil {
		bbox = &[4]float64{
			// correct BBox values is [opts.X, gp.curr.pageSize.H - opts.Y - opts.Rect.H, opts.X + opts.Rect.W, gp.curr.pageSize.H - opts.Y]
			// but if compress pdf through ghostscript result file can't open correctly in mac viewer, because mac viewer can't parse BBox value correctly
			// all other viewers parse BBox correctly (like Adobe Acrobat Reader, Chrome, even Internet Explorer)
			// that's why we need to set [0, 0, gp.curr.pageSize.W, gp.curr.pageSize.H]
			-gp.curr.pageSize.W * 2,
			-gp.curr.pageSize.H * 2,
			gp.curr.pageSize.W * 2,
			gp.curr.pageSize.H * 2,
			// Also, Chrome pdf viewer incorrectly recognize BBox value, that's why we need to set twice as much value
			// for every mask element will be displayed
		}
	}

	groupOpts := TransparencyXObjectGroupOptions{
		BBox:             *bbox,
		ExtGStateIndexes: opts.extGStateIndexes,
		XObjects:         []cacheContentImage{*image},
	}

	transparencyXObjectGroup, err := GetCachedTransparencyXObjectGroup(groupOpts, gp)
	if err != nil {
		return 0, err
	}

	sMaskOptions := SMaskOptions{
		Subtype:                       SMaskLuminositySubtype,
		TransparencyXObjectGroupIndex: transparencyXObjectGroup.Index,
	}
	sMask := GetCachedMask(sMaskOptions, gp)

	extGStateOpts := ExtGStateOptions{SMaskIndex: &sMask.Index}
	extGState, err := GetCachedExtGState(extGStateOpts, gp)
	if err != nil {
		return 0, err
	}

	return extGState.Index + 1, nil
}

func (gp *GoPdf) imageByHolder(img ImageHolder, opts ImageOptions) error {
	cacheImageIndex := -1

	for _, imgcache := range gp.curr.ImgCaches {
		if img.ID() == imgcache.Path {
			cacheImageIndex = imgcache.Index
			break
		}
	}

	if cacheImageIndex == -1 { //new image

		//create img object
		imgobj := new(ImageObj)
		if opts.Mask != nil {
			imgobj.SplittedMask = true
		}

		imgobj.init(func() *GoPdf {
			return gp
		})
		imgobj.setProtection(gp.protection())

		err := imgobj.SetImage(img)
		if err != nil {
			return err
		}

		if opts.Rect == nil {
			if opts.Rect, err = imgobj.getRect(); err != nil {
				return err
			}
		}

		err = imgobj.parse()
		if err != nil {
			return err
		}
		index := gp.addObj(imgobj)
		if gp.indexOfProcSet != -1 {
			//ยัดรูป
			procset := gp.pdfObjs[gp.indexOfProcSet].(*ProcSetObj)
			gp.getContent().AppendStreamImage(index, opts)
			procset.RelateXobjs = append(procset.RelateXobjs, RelateXobject{IndexOfObj: index})
			//เก็บข้อมูลรูปเอาไว้
			var imgcache ImageCache
			imgcache.Index = index
			imgcache.Path = img.ID()
			imgcache.Rect = opts.Rect
			gp.curr.ImgCaches[index] = imgcache
			gp.curr.CountOfImg++
		}

		if imgobj.haveSMask() {
			smaskObj, err := imgobj.createSMask()
			if err != nil {
				return err
			}
			imgobj.imginfo.smarkObjID = gp.addObj(smaskObj)
		}

		if imgobj.isColspaceIndexed() {
			dRGB, err := imgobj.createDeviceRGB()
			if err != nil {
				return err
			}
			dRGB.getRoot = func() *GoPdf {
				return gp
			}
			imgobj.imginfo.deviceRGBObjID = gp.addObj(dRGB)
		}

	} else { //same img
		if opts.Rect == nil {
			opts.Rect = gp.curr.ImgCaches[cacheImageIndex].Rect
		}

		gp.getContent().AppendStreamImage(cacheImageIndex, opts)
	}
	return nil
}

// Image : draw image
func (gp *GoPdf) Image(picPath string, x float64, y float64, rect *Rect) error {
	gp.UnitsToPointsVar(&x, &y)
	rect = rect.UnitsToPoints(gp.config.Unit)
	imgh, err := ImageHolderByPath(picPath)
	if err != nil {
		return err
	}

	imageOptions := ImageOptions{
		X:    x,
		Y:    y,
		Rect: rect,
	}

	return gp.imageByHolder(imgh, imageOptions)
}

func (gp *GoPdf) ImageFrom(img image.Image, x float64, y float64, rect *Rect) error {
	return gp.ImageFromWithOption(img, ImageFromOption{
		Format: "png",
		X:      x,
		Y:      y,
		Rect:   rect,
	})
}

func (gp *GoPdf) ImageFromWithOption(img image.Image, opts ImageFromOption) error {
	if img == nil {
		return newErr("Invalid image")
	}

	gp.UnitsToPointsVar(&opts.X, &opts.Y)
	opts.Rect = opts.Rect.UnitsToPoints(gp.config.Unit)
	r, w := io.Pipe()
	go func() {
		bw := bufio.NewWriter(w)
		var err error
		switch opts.Format {
		case "png":
			err = png.Encode(bw, img)
		case "jpeg":
			err = jpeg.Encode(bw, img, nil)
		}

		bw.Flush()
		if err != nil {
			w.CloseWithError(err)
		} else {
			w.Close()
		}
	}()

	imgh, err := ImageHolderByReader(bufio.NewReader(r))
	if err != nil {
		return err
	}

	imageOptions := ImageOptions{
		X:    opts.X,
		Y:    opts.Y,
		Rect: opts.Rect,
	}

	return gp.imageByHolder(imgh, imageOptions)
}
