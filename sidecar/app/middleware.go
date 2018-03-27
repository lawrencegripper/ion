package app

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const secret string = "secret"

//AddAuth enforces a shared secret between client and server for authentication
func AddAuth(secretHash string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sec := r.Header.Get(secret)
			err := CompareHash(sec, secretHash)
			if err != nil {
				respondWithError(err, http.StatusUnauthorized, w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

//AddIdentity adds an identity header to each requests
func AddIdentity(id string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("request-id", id)
			next.ServeHTTP(w, r)
		})
	}
}

//AddLog logs each request (Warning - performance impact)
func AddLog(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debugf("request received: %+v", r)
			next.ServeHTTP(w, r)
		})
	}
}

//getResourceID extracts the blob resource ID from the HTTP request
func getResourceID(r *http.Request) (string, error) {
	resPath := r.URL.Query().Get("res")
	reqID := r.Header.Get("request-id")
	if reqID == "" || resPath == "" {
		return "", fmt.Errorf("empty or invalid resource path or ID")
	}
	return fmt.Sprintf("%s/%s", reqID, resPath), nil
}
