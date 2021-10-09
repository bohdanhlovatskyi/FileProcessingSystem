package handlers

import (
	"encoding/json"
	"fmt"
	"io"
)

type Entry struct {
	File            string `json:"file"`
	Path            string `json:"path"`
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
