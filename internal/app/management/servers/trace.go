package servers

import (
	"context"
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/mongodb"
	"github.com/lawrencegripper/ion/internal/app/management/types"
	"github.com/lawrencegripper/ion/internal/pkg/management/trace"
)

//Check at compile time if we implement the interface
var _ trace.TraceServiceServer = (*TraceServer)(nil)

//NewTraceServer Create a new instance of a Trace management server
func NewTraceServer(config *types.Configuration) (*TraceServer, error) {
	traceServer := TraceServer{}

	mongoConnection, err := mongodb.NewMongoDB(&mongodb.Config{
		Collection: config.MongoDBCollection,
		Enabled:    true,
		Name:       config.MongoDBName,
		Password:   config.MongoDBPassword,
		Port:       config.MongoDBPort,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed connecting to mongo: %+v", err)
	}

	traceServer.mongoConnection = mongoConnection

	return &traceServer, nil
}

//TraceServer is an instance of a Trace management server
type TraceServer struct {
	mongoConnection *mongodb.MongoDB
}

//GetFlow returns json data from the meta store by correlationid
func (t *TraceServer) GetFlow(ctx context.Context, request *trace.GetFlowRequest) (*trace.GetFlowResponse, error) {
	json, err := t.mongoConnection.GetJSONDataByCorrelationID(request.CorrelationID)

	if err != nil {
		return nil, err
	}

	return &trace.GetFlowResponse{
		FlowJSON: *json,
	}, nil
}
