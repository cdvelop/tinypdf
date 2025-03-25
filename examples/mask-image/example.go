package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/cdvelop/tinypdf"
)

var resourcesPath string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	resourcesPath = filepath.Join(cwd, "test/res")
}

func main() {
	pdf := tinypdf.GoPdf{}

	pdf.Start(tinypdf.Config{PageSize: *tinypdf.PageSizeA4})
	pdf.AddPage()

	if err := pdf.AddTTFFont("loma", resourcesPath+"/LiberationSerif-Regular.ttf"); err != nil {
		log.Panic(err.Error())
	}

	if err := pdf.SetFont("loma", "", 14); err != nil {
		log.Panic(err.Error())
	}

	//image bytes
	b, err := os.ReadFile(resourcesPath + "/gopher01.jpg")
	if err != nil {
		log.Panic(err.Error())
	}

	imgH1, err := tinypdf.ImageHolderByBytes(b)
	if err != nil {
		log.Panic(err.Error())
	}
	if err := pdf.ImageByHolder(imgH1, 200, 250, nil); err != nil {
		log.Panic(err.Error())
	}

	//image io.Reader
	file, err := os.Open(resourcesPath + "/chilli.jpg")
	if err != nil {
		log.Panic(err.Error())
	}

	imgH2, err := tinypdf.ImageHolderByReader(file)
	if err != nil {
		log.Panic(err.Error())
	}

	maskHolder, err := tinypdf.ImageHolderByPath(resourcesPath + "/mask.png")
	if err != nil {
		log.Panic(err.Error())
	}

	maskOpts := tinypdf.MaskOptions{
		Holder: maskHolder,
		ImageOptions: tinypdf.ImageOptions{
			X: 0,
			Y: 0,
			Rect: &tinypdf.Rect{
				W: 300,
				H: 300,
			},
		},
	}

	transparency, err := tinypdf.NewTransparency(0.5, "")
	if err != nil {
		log.Panic(err.Error())
	}

	imOpts := tinypdf.ImageOptions{
		X:            0,
		Y:            0,
		Mask:         &maskOpts,
		Transparency: &transparency,
		Rect:         &tinypdf.Rect{W: 400, H: 400},
	}
	if err := pdf.ImageByHolderWithOptions(imgH2, imOpts); err != nil {
		log.Panic(err.Error())
	}

	pdf.SetCompressLevel(0)
	if err := pdf.WritePdf("image.pdf"); err != nil {
		log.Panic(err.Error())
	}
}
