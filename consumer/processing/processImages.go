package processing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/nfnt/resize"
)

type Processor struct {
	L log.Logger
}

type Entry struct {
	File            string `json:"file"`
	Path            string `json:"path"`
	Size            int64  `json:"size"`
	ContentType     string `json:"content-type"`
	TimeOfUploading string `json:"uploaded"`
	ID              int    `json:"id"`
}

func (e *Entry) FromJSON(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(e)
}

func (p *Processor) ProcessFile(descr []byte) error {
	var f Entry
	err := f.FromJSON(bytes.NewReader(descr))
	if err != nil {
		p.L.Printf("could not unmarshal json descriptor of file: %v\n", err)
		return err
	}

	file, err := os.Open(f.Path)
	if err != nil {
		p.L.Printf("could not open the file: %v\n", err)
		return err
	}

	decoder := map[string]func(io.Reader) (image.Image, error){
		"image/png":  png.Decode,
		"image/jpeg": jpeg.Decode,
	}

	// decode jpeg  image.Image
	d, ok := decoder[f.ContentType]
	if !ok {
		p.L.Printf("bad file format")
		return fmt.Errorf("bad file format")
	}
	img, err := d(file)
	if err != nil {
		p.L.Printf("could not decode the image: %v\n", err)
		return err
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(1000, 0, img, resize.Lanczos3)

	file, _ = os.OpenFile(f.Path, os.O_WRONLY, os.ModeAppend)

	// write new image to file
	err = png.Encode(file, m)
	if err != nil {
		p.L.Printf("coould not encode the image: %v\n", err)
		return err
	} else {
		p.L.Println("resized the image successfully")
	}

	return nil
}
