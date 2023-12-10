package handler

import (
	"fmt"
	"io"
	"net/http"
)

func Propmt(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	w.WriteHeader(200)
	fmt.Fprint(w, "Your query: "+string(body))
}
