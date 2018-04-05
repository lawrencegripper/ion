package mongodb

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/lawrencegripper/ion/sidecar/types"
	mongo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Config used to setup a MongoDB metastore provider
type Config struct {
	Name       string `description:"MongoDB database name"`
	Password   string `description:"MongoDB database password"`
	Collection string `description:"MongoDB database collection to use"`
	Port       int    `description:"MongoDB server port"`
}

//MongoDB handles the connection to an external Mongo database
type MongoDB struct {
	Session    *mongo.Session
	Collection *mongo.Collection
}

//NewMongoDB creates a new MongoDB object
func NewMongoDB(config *Config) (*MongoDB, error) {
	//TODO: Add timeout

	dialInfo := &mongo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s.documents.azure.com:%d", config.Name, config.Port)},
		Timeout:  60 * time.Second,
		Database: config.Name,
		Username: config.Name,
		Password: config.Password,
		DialServer: func(addr *mongo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		},
	}

	session, err := mongo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, fmt.Errorf("can't connect to mongo, go error %v", err)
	}

	session.SetSafe(&mongo.Safe{})

	col := session.DB(config.Name).C(config.Collection)

	MongoDB := &MongoDB{
		Session:    session,
		Collection: col,
	}

	return MongoDB, nil
}

//GetByID returns a single document matching a given document ID
func (db *MongoDB) GetByID(id string) (*types.Metadata, error) {
	metadata := types.Metadata{}
	err := db.Collection.Find(bson.M{"id": id}).One(&metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to get document with ID %s, error: %+v", id, err)
	}
	return &metadata, nil
}

//Create creats a new metadata document
func (db *MongoDB) Create(metadata *types.Metadata) error {
	b, err := json.Marshal(*metadata)
	if err != nil {
		return fmt.Errorf("error serializing JSON document: %+v", err)
	}
	var bsonDocument interface{}
	err = bson.UnmarshalJSON(b, &bsonDocument)
	if err != nil {
		return fmt.Errorf("error unmarshalling into BSON: %+v", err)
	}
	selector := bson.M{"id": metadata.ExecutionID}
	update := bson.M{"$set": metadata}
	_, err = db.Collection.Upsert(selector, update)
	if err != nil {
		return fmt.Errorf("error creates document: %+v", err)
	}
	return nil
}

//Append data to an existing metadata document
func (db *MongoDB) Append(id string, kvps []types.KeyValuePair) error {
	b, err := json.Marshal(kvps)
	if err != nil {
		return fmt.Errorf("error serializing JSON document: %+v", err)
	}
	var bsonDocument interface{}
	err = bson.UnmarshalJSON(b, &bsonDocument)
	if err != nil {
		return fmt.Errorf("error unmarshalling into BSON: %+v", err)
	}
	selector := bson.M{"id": id}
	for _, kvp := range kvps {
		update := bson.M{"$push": bson.M{"data": kvp}}
		err = db.Collection.Update(selector, update)
		if err != nil {
			return fmt.Errorf("error updating document: %+v", err)
		}
	}
	return nil
}

//Close cleans up the connection to Mongo
func (db *MongoDB) Close() {
	defer db.Session.Close()
}
