package app

import (
	"net/http"

	"github.com/lawrencegripper/ion/sidecar/types"
	log "github.com/sirupsen/logrus"
)

// cSpell:ignore debugf, logrus

const secret string = "secret"

//AddAuth enforces a shared secret between client and server for authentication
func AddAuth(secretHash string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sec := r.Header.Get(secret)
			err := types.CompareHash(sec, secretHash)
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
