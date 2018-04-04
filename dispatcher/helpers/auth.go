package helpers

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/lawrencegripper/ion/dispatcher/types"
	"github.com/sirupsen/logrus"
)

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func newServicePrincipalTokenFromCredentials(c *types.Configuration, scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		panic(err)
	}
	return adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, scope)
}

// GetAzureADAuthorizer return an authorizor for Azure SP
func GetAzureADAuthorizer(c *types.Configuration, azureEndpoint string) autorest.Authorizer {
	spt, err := newServicePrincipalTokenFromCredentials(c, azureEndpoint)
	if err != nil {
		logrus.Panicf("Failed to create authorizer: %v", err)
	}
	auth := autorest.NewBearerAuthorizer(spt)
	return auth
}
