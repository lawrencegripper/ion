package app

import (
	"fmt"
	"net/http"

	"github.com/lawrencegripper/mlops/sidecar/types"
	log "github.com/sirupsen/logrus"
)

const secret string = "secret"

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

//AddResolver extracts the blob resource ID from the request and invokes a resolver to
//construct a valid HTTP request to proxy.
func AddResolver(resolver types.Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resID, err := getResourceID(r)
			if err != nil {
				respondWithError(err, http.StatusBadRequest, w)
				return
			}
			r, err = resolver(resID, r)
			if err != nil {
				respondWithError(err, http.StatusBadRequest, w)
				return
			}
			r.RequestURI = "" // This is a hack to bypass this issue: https://github.com/vulcand/oxy/issues/57
			if err != nil {
				respondWithError(fmt.Errorf("failed to parse resource uri ('?res=resourceName') from query string"), http.StatusBadRequest, w)
				return
			}
			w.Header().Set("resource-id", resID)
			next.ServeHTTP(w, r)
		})
	}
}

//AddProxy forwards the request to another destination and invokes a post proxy function
func AddProxy(proxy types.BlobProxy, afterFunc func(w *StatusCodeResponseWriter)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := NewStatusCodeResponseWriter(w)
		proxy.ServeHTTP(rw, r)
		defer func() {
			if afterFunc != nil {
				afterFunc(rw)
			}
		}()
	})
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
