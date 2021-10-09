package handlers

import (
	"fmt"
	"net/http"
	"os"
)

func (u *Uploader) Display(w http.ResponseWriter, r *http.Request) {
	wd, _ := os.Getwd()
	http.ServeFile(w, r, fmt.Sprintf("%s/front/upload.html", wd))
}
