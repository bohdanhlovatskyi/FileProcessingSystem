package handlers

import (
	"bytes"
	"log"
	"os"

	"github.com/streadway/amqp"
)

const (
	DIR      = "uploaded_files"
	AMPQ_URL = "amqp://guest:guest@message-broker:5672"
)

func init() {
	os.Mkdir(DIR, 0777)
}

type Uploader struct {
	L           *log.Logger
	DataStorage []*Entry
	// sendConn    *amqp.Connection
	SendChan *amqp.Channel
}

func (u *Uploader) InitRabitMQ() error {
	// establishes connection
	conn, err := amqp.Dial(AMPQ_URL)
	if err != nil {
		u.L.Printf("could not set connection with the queue: %v\n", err)
		return err
	}
	// defer conn.Close()

	// opens a channel over the openned connection
	ch, err := conn.Channel()
	if err != nil {
		u.L.Printf("could not open a channel to communicate with queue: %v\n", err)
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
		u.L.Printf("could not create a queue over the openned channel: %v\n", err)
		return err
	}

	// u.sendConn = conn
	u.SendChan = ch

	return nil
}

func (u *Uploader) publishToQueue(fileID int) error {

	// prepare the  data
	var buf bytes.Buffer
	err := u.DataStorage[fileID].ToJSON(&buf)
	if err != nil {
		u.L.Printf("could not serialise data into json: %v\n", err)
		return err
	}
	// TODO: not sure that way of passing contenttype is ok here
	if u.SendChan != nil {
		err = u.SendChan.Publish(
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
		u.L.Printf("could not publish the message: %v\n", err)
		return err
	}

	if u.SendChan != nil {
		u.L.Print("Successfully published the message to the queue")
	}

	return nil
}
