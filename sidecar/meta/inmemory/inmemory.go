package inmemory

import (
	"fmt"

	"github.com/lawrencegripper/ion/sidecar/types"
)

//Config used to setup a InMemoryMetaProvider metastore provider
type Config struct {
	Initial map[string][]types.MetaDoc
}

//InMemoryMetaProvider is an in-memory implementation of a metastore
type InMemoryMetaProvider struct {
	docs map[string][]types.MetaDoc
}

//NewInMemoryMetaProvider returns a new InMemoryMetaProvider object
func NewInMemoryMetaProvider(config *Config) *InMemoryMetaProvider {
	if config.Initial == nil {
		config.Initial = make(map[string][]types.MetaDoc)
	}
	return &InMemoryMetaProvider{
		docs: config.Initial,
	}
}

//GetMetaDocByID retrieves a document given a document ID
func (p *InMemoryMetaProvider) GetMetaDocByID(docID string) (*types.MetaDoc, error) {
	for _, v := range p.docs {
		for _, doc := range v {
			if doc.ID == docID {
				return &doc, nil
			}
		}
	}
	return nil, fmt.Errorf("no document with a matching ID '%s' found in metastore", docID)
}

//GetMetaDocAll retrieves all documents stored under a correlation ID
func (p *InMemoryMetaProvider) GetMetaDocAll(correlationID string) ([]types.MetaDoc, error) {
	d, exists := p.docs[correlationID]
	if !exists {
		return nil, fmt.Errorf("no documents with a matching correlation ID '%s' found in metastore", correlationID)
	}
	return d, nil
}

//AddOrUpdateMetaDoc creates or updates a metadata document
func (p *InMemoryMetaProvider) AddOrUpdateMetaDoc(doc *types.MetaDoc) error {
	docs, exists := p.docs[doc.CorrelationID]
	if !exists {
		return fmt.Errorf("no documents with a matching correlation ID '%s' found in metastore", doc.CorrelationID)
	}
	for i, d := range docs {
		if d.ID == doc.ID {
			p.docs[doc.CorrelationID][i] = *doc
			return nil
		}
	}
	p.docs[doc.CorrelationID] = append(p.docs[doc.CorrelationID], *doc)
	return nil
}

//Close cleans up any external resources
func (p *InMemoryMetaProvider) Close() {
}
