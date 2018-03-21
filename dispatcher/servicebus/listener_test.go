package servicebus

import (
	"testing"
)

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestAmqpConnectionStringGeneration(t *testing.T) {
	const expectedConnString = "amqps://kn:kv@namespace.servicebus.windows.net"
	connString := getAmqpConnectionString("kn", "kv", "namespace")
	if connString != expectedConnString {
		t.Logf("Got: %s Expected: %s", connString, expectedConnString)
		t.Fail()
	}
}

func TestAmqpConnectionStringEncoding(t *testing.T) {
	const expectedConnString = `amqps://kn:bin%28%29@namespace.servicebus.windows.net`
	connString := getAmqpConnectionString("kn", "bin()", "namespace")
	if connString != expectedConnString {
		t.Logf("Got: %s Expected: %s", connString, expectedConnString)
		t.Fail()
	}
}

func TestGetSubscriptionName(t *testing.T) {
	const expected = `eventname_modulename`
	actual := getSubscriptionName("eventName", "moduleName")
	if actual != expected {
		t.Logf("Got: %s Expected: %s", actual, expected)
		t.Fail()
	}
}

func TestGetSubscriptionAmqpPath(t *testing.T) {
	const expected = `/exampleevent/subscriptions/exampleevent_modulename`
	actual := getSubscriptionAmqpPath("exampleEvent", "moduleName")
	if actual != expected {
		t.Logf("Got: %s Expected: %s", actual, expected)
		t.Fail()
	}
}
