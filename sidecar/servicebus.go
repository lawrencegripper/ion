package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"k8s.io/kubernetes/third_party/forked/golang/template"
)

//ServiceBus handles the connection to an external Service Bus
type ServiceBus struct {
	URL string
	SAS string
	SKN string
}

//NewServiceBus creates a new Service Bus object
func NewServiceBus(namespace, topic, key, skn string) (*ServiceBus, error) {
	sb := &ServiceBus{
		URL: fmt.Sprintf("https://%s.servicebus.windows.net/%s/messages", namespace, topic),
		SAS: key,
		SKN: skn,
	}
	//TODO: validate connection for fast failure
	return sb, nil
}

//PublishEvent publishes an event onto a Service Bus topic
func (s *ServiceBus) PublishEvent(e Event) (int, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error publishing event %+v", err)
	}
	req, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generateSAS(s.URL, s.SKN, s.SAS))

	//TODO: optimize
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error publishing event %+v", err)
	}

	switch res.StatusCode {
	case http.StatusCreated:
		return res.StatusCode, nil
	case http.StatusBadRequest:
		return res.StatusCode, fmt.Errorf("bad request")
	case http.StatusUnauthorized:
		return res.StatusCode, fmt.Errorf("authorization failure")
	case http.StatusForbidden:
		return res.StatusCode, fmt.Errorf("quota exceeded or message to large")
	case http.StatusGone:
		return res.StatusCode, fmt.Errorf("specified queue or topic does not exist")
	case http.StatusInternalServerError:
		return res.StatusCode, fmt.Errorf("internal error")
	default:
		return res.StatusCode, fmt.Errorf("unknown status code")
	}
}

//Close cleans up any outstanding connections to Service Bus
func (s *ServiceBus) Close() {
}

//generateSAS builds a SAS token for use as a HTTP header
// nolint: errcheck
func generateSAS(uri, skn, key string) string {
	encoded := template.URLQueryEscaper(uri)
	now := time.Now().Unix()
	week := 60 * 60 * 24 * 7
	ts := now + int64(week)
	signature := encoded + "\n" + strconv.Itoa(int(ts))
	k := []byte(key)
	hmac := hmac.New(sha256.New, k)
	hmac.Write([]byte(signature))
	hmacString := template.URLQueryEscaper(base64.StdEncoding.EncodeToString(hmac.Sum(nil)))

	result := "SharedAccessSignature sr=" + encoded + "&sig=" +
		hmacString + "&se=" + strconv.Itoa(int(ts)) + "&skn=" + skn
	return result
}
