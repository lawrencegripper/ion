package providers

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/go-autorest/autorest"

	"github.com/Azure/azure-sdk-for-go/services/batch/2017-09-01.6.0/batch"
)

func getPool(ctx context.Context, batchBaseURL, poolID string, auth autorest.Authorizer) (*batch.PoolClient, error) {
	poolClient := batch.NewPoolClientWithBaseURI(batchBaseURL)
	poolClient.Authorizer = auth
	poolClient.RetryAttempts = 0

	pool, err := poolClient.Get(ctx, poolID, "*", "", nil, nil, nil, nil, "", "", nil, nil)

	// If we observe an error which isn't related to the pool not existing panic.
	// 404 is expected if this is first run.
	if err != nil && pool.Response.Response == nil {
		log.WithError(err).Error("Failed to get pool. nil response %v", poolID)
		return nil, err
	} else if err != nil && pool.StatusCode == 404 {
		log.WithError(err).Error("Pool doesn't exist 404 received PoolID: %v", poolID)
		return nil, err
	} else if err != nil {
		log.WithError(err).Error("Failed to get pool. Response:%v", pool.Response)
		return nil, err
	}

	if pool.State == batch.PoolStateActive {
		log.Println("Pool active and running...")
		return &poolClient, nil
	}
	return nil, fmt.Errorf("Pool not in active state: %v", pool.State)
}

func createOrGetJob(ctx context.Context, batchBaseURL, jobID, poolID string, auth autorest.Authorizer) (*batch.JobClient, error) {
	jobClient := batch.NewJobClientWithBaseURI(batchBaseURL)
	jobClient.Authorizer = auth
	// check if job exists already
	currentJob, err := jobClient.Get(ctx, jobID, "", "", nil, nil, nil, nil, "", "", nil, nil)

	if err == nil && currentJob.State == batch.JobStateActive {
		log.Println("Wrapper job already exists...")
		return &jobClient, nil
	} else if currentJob.Response.StatusCode == 404 {

		log.Println("Wrapper job missing... creating...")
		wrapperJob := batch.JobAddParameter{
			ID: &jobID,
			PoolInfo: &batch.PoolInformation{
				PoolID: &poolID,
			},
		}

		res, err := jobClient.Add(ctx, wrapperJob, nil, nil, nil, nil)

		log.WithField("jobReponse", res).Debug("created job")

		if err != nil {
			return nil, err
		}
		return &jobClient, nil

	} else if currentJob.State == batch.JobStateDeleting {
		log.Info("Job is being deleted... Waiting then will retry")
		time.Sleep(time.Minute)
		return createOrGetJob(ctx, batchBaseURL, jobID, poolID, auth)
	}

	return nil, err
}

func getBatchBaseURL(batchAccountName, batchAccountLocation string) string {
	return fmt.Sprintf("https://%s.%s.batch.azure.com", batchAccountName, batchAccountLocation)
}
