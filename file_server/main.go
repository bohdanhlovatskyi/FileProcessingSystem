package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

const (
	DIR = "uploaded_files"
)

func init() {
	os.Mkdir(DIR, 0777)
}

func main() {
	l := log.New(os.Stdout, "file-server", log.LstdFlags)

	upl := &Uploader{
		l: l,
	}

	sm := mux.NewRouter()
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/upload", upl.Display)

	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/upload", upl.AddHandler)

	s := &http.Server{
		Addr:         ":8080",
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// run the server in separate goroutine, so to
	// implement graceful shutdown
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// consume message
	sig := <-sigChan
	l.Println("received terminate, graceful shutdown", sig)

	tc, cf := context.WithTimeout(context.Background(), 20*time.Second)
	cf() // calling the context function
	s.Shutdown(tc)
}

type Uploader struct {
	l           *log.Logger
	DataStorage []*Entry
}

var templates = template.Must(template.ParseFiles("front/upload.html"))

func display(w http.ResponseWriter, page string, data interface{}) {
	templates.ExecuteTemplate(w, page+".html", data)
}

func (u *Uploader) Display(w http.ResponseWriter, r *http.Request) {
	display(w, "upload", nil)
}

type Entry struct {
	File            string
	Path            string
	Size            int64
	ContentType     string
	TimeOfUploading string
}

func (f *Entry) String() string {
	return fmt.Sprintf(
		"File(%s, path: %s, size: %d, content type: %s, time of uploading: %s)",
		f.File, f.Path, f.Size, f.ContentType, f.TimeOfUploading,
	)
}

func (u *Uploader) AddHandler(rw http.ResponseWriter, r *http.Request) {

	// parses the file from HTTP response
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		u.l.Printf("addHandler: could not form the file from input: %v\n", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// creates a file on the local storage to save the file
	filePath := filepath.Join(DIR, header.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		u.l.Printf("addHandler: could not create a file: %v\n", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// copies the contents of the file
	if _, err := io.Copy(dst, file); err != nil {
		u.l.Printf("addHandler: could not save the file: %v\n", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// creates an entry
	entry := &Entry{
		File:            header.Filename,
		Path:            filePath,
		Size:            header.Size,
		ContentType:     header.Header["Content-Type"][0],
		TimeOfUploading: time.Now().String(),
	}

	u.DataStorage = append(u.DataStorage, entry)

	fmt.Fprint(rw, entry.String())
}
