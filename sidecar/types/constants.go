package types

//ContentType represents the 'Content-Type' HTTP header key
const ContentType string = "Content-Type"

//ContentTypeApplicationJSON represents the 'Content-Type' HTTP header value "application/json"
const ContentTypeApplicationJSON string = "application/json"

//MetaProviderMongoDB is a lowercase name for the MongoDB metadata provider
const MetaProviderMongoDB string = "mongodb"

//MetaProviderInMemory is a lowercase name for the in-memory metadata provider
const MetaProviderInMemory string = "inmemory"

//BlobProviderAzureStorage is a lowercase name for the azure blob storage blob provider
const BlobProviderAzureStorage string = "azureblob"

//BlobProviderFileSystem is a lowercase name for the file system blob provider
const BlobProviderFileSystem string = "filesystem"

//EventProviderServiceBus is a lowercase name for the file system blob provider
const EventProviderServiceBus string = "servicebus"
