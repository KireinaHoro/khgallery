package main

import (
	"fmt"
	"golang.org/x/image/draw"
	"html/template"
	"image"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TODO: scan recursively instead of assuming flat structure
const scanDir = "/Users/jsteward/work/jsteward.moe/content/images/gallery/"
const thumbnailsDirName = "thumbnails/"
const thumbnailsDir = scanDir + thumbnailsDirName
const deploymentHref = "/images/gallery/"
const thumbnailWidthRegular = 500
const thumbnailWidthPanorama = 2000
const panoramaRatio = 5 // width >= 5 * height
const templateName = "gallery.md"

var glTmpl = template.Must(template.New(templateName).ParseFiles(templateName))

// A PhotoInfo contains all metadata of a photo needed to generate the photoswipe/isotope <div> in the gallery page.
type PhotoInfo struct {
	CollectionName string
	Filename       string
	Width          int
	Height         int
	IsPanorama     bool
}

type TemplateCtx struct {
	Data          Gallery
	Slug          func(string) string
	DeployHref    string
	ThumbnailsDir string
}

type Gallery struct {
	PiArr []PhotoInfo
	Name  string
	Date  string
}

func slug(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

func (gl *Gallery) pushPhoto(pi *PhotoInfo) {
	gl.PiArr = append(gl.PiArr, *pi)
}

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
	gl := Gallery{
		Name: "Test Gallery",
		Date: time.Now().Format(time.DateTime),
	}
	piChan := make(chan *PhotoInfo)
	collectChan := make(chan interface{})
	// collect PhotoInfo
	go func() {
		for pi := range piChan {
			log.Printf("Found image %+v", pi)
			gl.pushPhoto(pi)
		}
		collectChan <- struct{}{}
	}()

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
				piChan <- pi
			}
		}(fp)
	}
	wg.Wait()
	close(piChan)
	<-collectChan

	// shuffle the photos
	// TODO: sort temporally (shuffle should be done in js)
	rand.Shuffle(len(gl.PiArr), func(i, j int) {
		gl.PiArr[i], gl.PiArr[j] = gl.PiArr[j], gl.PiArr[i]
	})

	ctx := TemplateCtx{
		Data:          gl,
		Slug:          slug,
		DeployHref:    deploymentHref,
		ThumbnailsDir: thumbnailsDirName,
	}

	outName := filepath.Join(scanDir, "..", "..", "Gallery", "test.md")
	out, err := os.Create(outName)
	if err != nil {
		log.Fatalf("failed to open output file: %s", err)
	}
	defer func() {
		err = out.Close()
		if err != nil {
			log.Fatalf("failed to close output file: %s", err)
		}
	}()
	err = glTmpl.Execute(out, ctx)
	if err != nil {
		log.Fatalf("failed to generate gallery markdown page: %s", err)
	}
	log.Printf("Written output %s", outName)
}
