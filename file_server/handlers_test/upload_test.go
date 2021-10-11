package handlers_test

import (
	"bytes"
	"encoding/json"
	"files/file_server/handlers"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	test_img_url = "./test_png.png"
	url          = "localhost:8080/upload"
)

func TestAddHandler(t *testing.T) {
	upl := &handlers.Uploader{L: log.New(os.Stdout, "file-server", log.LstdFlags)}

	var b bytes.Buffer
	var fw io.Writer
	var err error
	w := multipart.NewWriter(&b)

	// prepares MultipartForm
	file := mustOpen(test_img_url)
	if fw, err = w.CreateFormFile("file", file.Name()); err != nil {
		t.Errorf("Error creating writer: %v", err)
	}
	if _, err := io.Copy(fw, file); err != nil {
		t.Errorf("Error with io.Copy: %v", err)
	}
	w.Close()

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/upload", &b)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(upl.AddHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// decodes the response json
	e := &handlers.Entry{}
	d := json.NewDecoder(bytes.NewBufferString(rr.Body.String()))
	err = d.Decode(e)
	if err != nil {
		t.Fatal("could not parse the json response\n")
	}

	// creating expected struct to be received
	expected := &handlers.Entry{
		File: "test_png.png",
		Path: "uploaded_files/test_png.png",
		Size: 83054,
	}

	// --------- actual testing ------
	if expected.File != e.File {
		t.Errorf("handler returned unexpected file name: got %v want %v",
			e.File, expected.File)
	}

	if expected.Path != e.Path {
		t.Errorf("handler returned unexpected saving path: got %v want %v",
			e.Path, expected.Path)
	}

	if expected.Size != e.Size {
		t.Errorf("handler returned unexpected file size: got %v want %v",
			e.Size, expected.Size)
	}

	// clean up
	os.RemoveAll("uploaded_files")
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		pwd, _ := os.Getwd()
		fmt.Println("PWD: ", pwd)
		panic(err)
	}
	return r
}
