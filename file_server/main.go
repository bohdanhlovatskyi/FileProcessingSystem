package main

import (
	"context"
	"files/file_server/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	l := log.New(os.Stdout, "file-server", log.LstdFlags)

	upl := &handlers.Uploader{
		L: l,
	}

	err := upl.InitRabitMQ()
	// TODO: this is obviously not the best way to go
	defer upl.SendChan.Close()
	//defer upl.sendConn.Close()

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
