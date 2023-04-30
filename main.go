package main

import (
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// A PhotoInfo contains all metadata of a photo needed to generate the photoswipe/isotope <div> in the gallery page.
type PhotoInfo struct {
	CollectionName string
	Filename       string
	Width          int
	Height         int
	IsPanorama     bool
}

// TODO: scan recursively instead of assuming flat structure
const scanDir string = "/Users/jsteward/Pictures/test-gallery-client"
const thumbnailsDir string = scanDir + "/thumbnails"
const thumbnailWidthRegular int = 500
const thumbnailWidthPanorama int = 2000
const panoramaRatio int = 5 // width >= 5 * height

// For a single image, generate its thumbnail and read the related metadata to fill the `PhotoInfo` struct.  Only I/O
// related operations should be performed here; templating to create the gallery HTML is done in `galleryCodeGen`.
func doSingleImage(fp os.DirEntry) (_ *PhotoInfo, err error) {
	n := fp.Name()
	path := filepath.Join(scanDir, n)
	ext := filepath.Ext(n)
	if ext != ".jpg" {
		return nil, fmt.Errorf("unsupported extension name %s", ext)
	}
	reader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w; failed to open file", err)
	}
	defer func(reader *os.File) {
		err = reader.Close()
	}(reader)
	// generate thumbnail
	im, err := jpeg.Decode(reader)
	sz := im.Bounds().Size()
	if err != nil {
		return nil, fmt.Errorf("%w; failed to decode image", err)
	}
	out, err := os.Create(filepath.Join(thumbnailsDir, n))
	if err != nil {
		return nil, fmt.Errorf("%w; failed to create thumbnail file", err)
	}
	defer func(out *os.File) {
		err = out.Close()
	}(out)
	isPanorama := false
	thumbnailX := thumbnailWidthRegular
	if sz.X/sz.Y >= panoramaRatio {
		thumbnailX = thumbnailWidthPanorama
		isPanorama = true
	}
	thumbnailY := thumbnailX * sz.Y / sz.X
	// https://stackoverflow.com/a/67678654/5520728
	outIm := image.NewRGBA(image.Rect(0, 0, thumbnailX, thumbnailY))
	draw.BiLinear.Scale(outIm, outIm.Rect, im, im.Bounds(), draw.Over, nil)
	err = jpeg.Encode(out, outIm, nil)
	if err != nil {
		return nil, fmt.Errorf("%w; failed to encode thumbnail", err)
	}

	// fill in metadata
	var collectionName string
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		collectionName = "(default)"
	} else {
		dir, _ := filepath.Split(realPath)
		collectionName = filepath.Base(dir)
	}
	return &PhotoInfo{
		CollectionName: collectionName,
		Filename:       n,
		Width:          sz.X,
		Height:         sz.Y,
		IsPanorama:     isPanorama,
	}, nil
}

func galleryCodeGen() {

}

func main() {
	files, err := os.ReadDir(scanDir)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(thumbnailsDir, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create thumbnails directory: %s", err)
	}
	var wg sync.WaitGroup
	for _, fp := range files {
		if fp.IsDir() || filepath.Ext(fp.Name()) != ".jpg" {
			continue
		}
		go func(fp os.DirEntry) {
			wg.Add(1)
			defer wg.Done()
			pi, err := doSingleImage(fp)
			if err != nil {
				log.Printf("failed to process single image %s: %s", fp.Name(), err)
			} else {
				log.Printf("Found image %+v", pi)
			}
		}(fp)
	}
	wg.Wait()
}
