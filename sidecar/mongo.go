package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"

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
		return nil, fmt.Errorf("can't connect to mongo, go error %v\n", err)
	}

	session.SetSafe(&mongo.Safe{})

	col := session.DB(name).C(collection)

	mongoDB := &MongoDB{
		Session:    session,
		Collection: col,
	}

	return mongoDB, nil
}

//GetByID returns the document with a matching reference ID
func (db *MongoDB) GetByID(id string) (*Document, error) {
	result := Document{}
	err := db.Collection.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get document with ID %s, error: %+v", id, err)
	}
	return &result, nil
}

//Update appends a new entry to an existing document
func (db *MongoDB) Update(id string, entry Entry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("error updating document: %+v", err)
	}
	var bsonDocument interface{}
	err = bson.UnmarshalJSON(b, &bsonDocument)
	if err != nil {
		return fmt.Errorf("error updating document: %+v", err)
	}
	updateQuery := bson.M{"id": id}
	patch := bson.M{"$push": bson.M{"entries": bsonDocument}}
	err = db.Collection.Update(updateQuery, patch)
	if err != nil {
		return fmt.Errorf("error updating document: %+v", err)
	}
	return nil
}

//Close cleans up the connection to Mongo
func (db *MongoDB) Close() {
	defer db.Session.Close()
}
