package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/streadway/amqp"
)

const (
	DIR      = "uploaded_files"
	AMPQ_URL = "amqp://guest:guest@localhost:5672/"
)

func init() {
	os.Mkdir(DIR, 0777)
}

type Uploader struct {
	l           *log.Logger
	DataStorage []*Entry
	sendConn    *amqp.Connection
	sendChan    *amqp.Channel
}

func (u *Uploader) init_rabitMQ() error {
	// establishes connection
	conn, err := amqp.Dial(AMPQ_URL)
	if err != nil {
		u.l.Printf("could not set connection with the queue: %v\n", err)
		return err
	}
	// defer conn.Close()

	// opens a channel over the openned connection
	ch, err := conn.Channel()
	if err != nil {
		u.l.Printf("could not open a channel to communicate with queue: %v\n", err)
		return err
	}
	// defer ch.Close()

	// declares queue over the openned channel
	_, err = ch.QueueDeclare(
		"UploadedFiles",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		u.l.Printf("could not create a queue over the openned channel: %v\n", err)
		return err
	}

	u.sendConn = conn
	u.sendChan = ch

	return nil
}

func main() {
	l := log.New(os.Stdout, "file-server", log.LstdFlags)

	upl := &Uploader{
		l: l,
	}

	err := upl.init_rabitMQ()
	// TODO: this is obviously not the best way to go
	defer upl.sendChan.Close()
	defer upl.sendConn.Close()

	if err != nil {
		// TODO: not sure that we need to terminate it here
		// log.Fatal(err)
		l.Printf("could not connect to the rabbit mq\n")
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

var templates = template.Must(template.ParseFiles("front/upload.html"))

func display(w http.ResponseWriter, page string, data interface{}) {
	templates.ExecuteTemplate(w, page+".html", data)
}

func (u *Uploader) Display(w http.ResponseWriter, r *http.Request) {
	display(w, "upload", nil)
}

type Entry struct {
	File            string `json:"file"`
	Path            string `json:"-"`
	Size            int64  `json:"size"`
	ContentType     string `json:"content-type"`
	TimeOfUploading string `json:"uploaded"`
	ID              int    `json:"id"`
}

func (f *Entry) String() string {
	return fmt.Sprintf(
		"File(%s, path: %s, size: %d, content type: %s, time of uploading: %s)",
		f.File, f.Path, f.Size, f.ContentType, f.TimeOfUploading,
	)
}

func (f *Entry) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(f)
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
		ID:              len(u.DataStorage),
	}

	u.DataStorage = append(u.DataStorage, entry)

	err = u.publishToQueue(entry.ID)
	if err != nil {
		u.l.Print(err)
	}

	fmt.Fprint(rw, entry.String())
}

func (u *Uploader) publishToQueue(fileID int) error {

	// prepare the  data
	var buf bytes.Buffer
	err := u.DataStorage[fileID].ToJSON(&buf)
	if err != nil {
		u.l.Printf("could not serialise data into json: %v\n", err)
		return err
	}
	// TODO: not sure that way of passing contenttype is ok here
	if u.sendChan != nil {
		err = u.sendChan.Publish(
			"",
			"UploadedFiles",
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        buf.Bytes(),
			},
		)
	}

	if err != nil {
		u.l.Printf("could not publish the message: %v\n", err)
		return err
	}

	if u.sendChan != nil {
		u.l.Print("Successfully published the message to the queue")
	}

	return nil
}
