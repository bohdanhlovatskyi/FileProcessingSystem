package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func (u *Uploader) AddHandler(rw http.ResponseWriter, r *http.Request) {

	// parses the file from HTTP response
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		u.L.Printf("addHandler: could not form the file from input: %v\n", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// very simple validation, should be rewritten
	content_type := header.Header["Content-Type"][0]

	// the last one is for unit testing
	reg, _ := regexp.Compile("(png|jpeg|octet-stream)")
	if !reg.MatchString(content_type) {
		fmt.Println(header.Header["Content-Type"])
		http.Error(rw, "you can upload only images", http.StatusBadRequest)
		return
	}

	// creates a file on the local storage to save the file
	filePath := filepath.Join(DIR, header.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		u.L.Printf("addHandler: could not create a file: %v\n", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// copies the contents of the file
	if _, err := io.Copy(dst, file); err != nil {
		u.L.Printf("addHandler: could not save the file: %v\n", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// creates an entry
	entry := &Entry{
		File:            header.Filename,
		Path:            filePath,
		Size:            header.Size,
		ContentType:     content_type,
		TimeOfUploading: time.Now().String(),
		ID:              len(u.DataStorage),
	}

	u.DataStorage = append(u.DataStorage, entry)

	err = u.publishToQueue(entry.ID)
	if err != nil {
		u.L.Print(err)
	}

	fmt.Fprint(rw, entry.ToJSON(rw))
}
