package types

// Configuration for the application
type Configuration struct {
	Hostname            string
	LogLevel            string `short:"l" description:"Log level"`
	ModuleName          string `description:"Name of the module"`
	SubscribesToEvent   string `descriptions:"Event this modules subscribes to"`
	EventsPublished     string `descriptions:"Events this modules can publish"`
	ServiceBusNamespace string `descriptions:"Namespace to use for ServiceBus"`
	ResourceGroup       string `descriptions:"Azure ResourceGroup to use"`
	SubscriptionID      string `description:"SubscriptionID for Azure"`
	ClientID            string `description:"ClientID of Service Principal for Azure access"`
	ClientSecret        string `description:"Client Secrete of Service Principal for Azure access"`
	TenantID            string `description:"TentantID for Azure"`
	LogSensitiveConfig  bool   `description:"Print out sensitive config when logging"`
}

// RedactConfigSecrets strips sensitive data from the config
func RedactConfigSecrets(c Configuration) Configuration {
	if !c.LogSensitiveConfig {
		c.ClientID = "***********"
		c.ClientSecret = "***********"
		c.TenantID = "***********"
		c.SubscriptionID = "***********"
	}
	return c
}
