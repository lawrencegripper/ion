package servicebus

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lawrencegripper/ion/internal/pkg/common"
	"k8s.io/kubernetes/third_party/forked/golang/template"
)

// cSpell:ignore kubernetes, nolint

//Config to setup a ServiceBus event publisher
type Config struct {
	Enabled               bool   `description:"Enable Service Bus event provider"`
	Namespace             string `description:"ServiceBus namespace"`
	Topic                 string `description:"ServiceBus topic name"`
	Key                   string `description:"ServiceBus access key"`
	AuthorizationRuleName string `description:"ServiceBus authorization rule name"`
}

//ServiceBus handles the connection to an external Service Bus
type ServiceBus struct {
	PartialURL string `description:"This is the SB url without the topic name set. Replace %%TOPIC_PLACEHOLDER%% with the required topic before using"`
	Key        string
	SKN        string
}

const topicPlaceholderText = "%%TOPIC_PLACEHOLDER%%"

//NewServiceBus creates a new Service Bus object
func NewServiceBus(config *Config) (*ServiceBus, error) {
	sb := &ServiceBus{
		PartialURL: fmt.Sprintf("https://%s.servicebus.windows.net/%s/messages", config.Namespace, topicPlaceholderText),
		Key:        config.Key,
		SKN:        config.AuthorizationRuleName,
	}
	//TODO: validate connection for fast failure
	return sb, nil
}

//Publish publishes an event onto a Service Bus topic
func (s *ServiceBus) Publish(e common.Event) error {

	// generate a url for the correct topic for this event
	sbURL := strings.Replace(s.PartialURL, topicPlaceholderText, e.Type, -1)

	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("error publishing event %+v", err)
	}
	req, err := http.NewRequest(http.MethodPost, sbURL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generateSAS(sbURL, s.SKN, s.Key))

	//TODO: optimize
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error publishing event %+v", err)
	}

	switch res.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest:
		return fmt.Errorf("bad request")
	case http.StatusUnauthorized:
		return fmt.Errorf("authorization failure")
	case http.StatusForbidden:
		return fmt.Errorf("quota exceeded or message to large")
	case http.StatusGone:
		return fmt.Errorf("specified queue or topic does not exist")
	case http.StatusInternalServerError:
		return fmt.Errorf("internal error")
	default:
		return fmt.Errorf("unknown status code")
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
