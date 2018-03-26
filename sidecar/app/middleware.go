package app

import (
	"fmt"
	"net/http"
	"net/url"

	c "github.com/lawrencegripper/mlops/sidecar/common"
	log "github.com/sirupsen/logrus"
)

//StatusCodeResponseWriter is used to expose the HTTP status code for a ResponseWriter
type StatusCodeResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

//NewStatusCodeResponseWriter creates new StatusCodeResponseWriter
func NewStatusCodeResponseWriter(w http.ResponseWriter) *StatusCodeResponseWriter {
	return &StatusCodeResponseWriter{w, http.StatusOK}
}

//WriteHeader hijacks a ResponseWriter.WriteHeader call and stores the status code
func (w *StatusCodeResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

//AddAuth enforces a shared secret between client and server for authentication
func AddAuth(secretHash string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secret := r.Header.Get(secretHeaderKey)
			err := c.CompareHash(secret, secretHash)
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

//AddProxy proxies the request to another destination
func AddProxy(app *App, headers map[string]string,
	resolveFunc func(url string) (string, error),
	afterFunc func(res string, w StatusCodeResponseWriter)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resource := r.URL.Query().Get("res")
		id := r.Header.Get("request-id")
		if id == "" || resource == "" {
			respondWithError(fmt.Errorf("resource or id could not be found in request"), http.StatusBadRequest, w)
			return
		}
		resourceID := id + "/" + resource
		resourceURI, err := resolveFunc(resourceID)
		if err != nil {
			respondWithError(err, http.StatusBadRequest, w)
			return
		}

		for k, v := range headers {
			r.Header.Set(k, v)
		}
		r.URL, err = url.Parse(resourceURI)
		r.RequestURI = "" // This is a hack to bypass this issue: https://github.com/vulcand/oxy/issues/57
		if err != nil {
			respondWithError(fmt.Errorf("failed to parse resource uri ('?res=resourceName') from query string"), http.StatusBadRequest, w)
			return
		}
		rw := NewStatusCodeResponseWriter(w)
		app.Proxy.ServeHTTP(rw, r)
		defer afterFunc(resourceURI, *rw)
	})
}
