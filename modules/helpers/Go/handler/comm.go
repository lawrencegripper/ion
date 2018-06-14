package handler

import (
	"net/http"
	"os"
	"strconv"

	"github.com/lawrencegripper/ion/modules/helpers/Go/log"
)

var port int64 = 8080

func init() {
	p := os.Getenv("SIDECAR_PORT")
	if p != "" {
		p, err := strconv.ParseInt(p, 10, 32)
		if err == nil {
			port = p
		}
	}
}
func Ready() bool {
	r, err := http.Get("http://localhost:" + strconv.FormatInt(port, 10) + "/ready")

	if err == nil && r.StatusCode == http.StatusOK {
		return true
	}

	log.Fatal("Ready command failed")
	return false
}

func Commit() bool {
	r, err := http.Get("http://localhost:" + strconv.FormatInt(port, 10) + "/done")

	if err == nil && r.StatusCode == http.StatusOK {
		return true
	}

	log.Fatal("Done command failed")

	return false
}
