package processing

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"testing"
)

func TestProcessFile(t *testing.T) {
	p := &Processor{
		L: log.New(os.Stdout, "file-server", log.LstdFlags),
	}

	// mock file
	file := &Entry{
		File:        "test_png.png",
		Path:        "test_png.png",
		ContentType: "image/png",
		Size:        83054,
	}

	// encode the entry
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.Encode(file)

	err := p.ProcessFile(buf.Bytes())
	if err != nil {
		t.Errorf("could not process the file: %v\n", err)
	}
}
