package azure

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/lawrencegripper/mlops/sidecar/common"
	mongo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//MongoDB handles the connection to an external Mongo database
type MongoDB struct {
	Session    *mongo.Session
	Collection *mongo.Collection
}

//NewMongoDB creates a new MongoDB object
func NewMongoDB(name, password, collection string, port int) (*MongoDB, error) {
	dialInfo := &mongo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s.documents.azure.com:%d", name, port)},
		Timeout:  60 * time.Second,
		Database: name,
		Username: name,
		Password: password,
		DialServer: func(addr *mongo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		},
	}

	session, err := mongo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, fmt.Errorf("can't connect to mongo, go error %v", err)
	}

	session.SetSafe(&mongo.Safe{})

	col := session.DB(name).C(collection)

	mongoDB := &MongoDB{
		Session:    session,
		Collection: col,
	}

	return mongoDB, nil
}

//GetMetaDocByID returns a single document matching a given document ID
func (db *MongoDB) GetMetaDocByID(docID string) (*common.MetaDoc, error) {
	doc := common.MetaDoc{}
	err := db.Collection.Find(bson.M{"id": docID}).One(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to get document with ID %s, error: %+v", docID, err)
	}
	return &doc, nil
}

//GetMetaDocAll returns all the documents matching a given correlationID
func (db *MongoDB) GetMetaDocAll(correlationID string) ([]common.MetaDoc, error) {
	docs := []common.MetaDoc{}
	err := db.Collection.Find(bson.M{"correlationId": correlationID}).All(&docs)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents with correlation ID %s, error: %+v", correlationID, err)
	}
	return docs, nil
}

//AddOrUpdateMetaDoc appends a new entry to an existing document
func (db *MongoDB) AddOrUpdateMetaDoc(doc *common.MetaDoc) error {
	b, err := json.Marshal(*doc)
	if err != nil {
		return fmt.Errorf("error serializing JSON document: %+v", err)
	}
	var bsonDocument interface{}
	err = bson.UnmarshalJSON(b, &bsonDocument)
	if err != nil {
		return fmt.Errorf("error unmarshalling into BSON: %+v", err)
	}

	selector := bson.M{"id": doc.ID}
	update := bson.M{"$set": doc}
	db.Collection.Upsert(selector, update)
	if err != nil {
		return fmt.Errorf("error updating document: %+v", err)
	}
	return nil
}

//Close cleans up the connection to Mongo
func (db *MongoDB) Close() {
	defer db.Session.Close()
}
