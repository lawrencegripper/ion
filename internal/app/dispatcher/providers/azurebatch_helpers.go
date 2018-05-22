package providers

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
import (
	"fmt"
	"net/http"
	"time"

	"github.com/lawrencegripper/ion/internal/pkg/types"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/go-autorest/autorest"

	"github.com/Azure/azure-sdk-for-go/services/batch/2017-09-01.6.0/batch"
	"github.com/Azure/go-autorest/autorest/to"
)

func createOrGetPool(p *AzureBatch, auth autorest.Authorizer) {

	poolClient := batch.NewPoolClientWithBaseURI(getBatchBaseURL(p.batchConfig))
	poolClient.Authorizer = auth
	poolClient.RetryAttempts = 0
	p.poolClient = &poolClient
	log.Warningln("echo")

	pool, err := poolClient.Get(p.ctx, p.batchConfig.PoolID, "*", "", nil, nil, nil, nil, "", "", nil, nil)
	log.Warningln("echo")
	log.Info(pool)
	log.Info(err)

	// If we observe an error which isn't related to the pool not existing panic.
	// 404 is expected if this is first run.
	if err != nil && pool.Response.Response == nil {
		log.WithError(err).Panicf("Failed to get pool. nil response %v", p.batchConfig.PoolID)
	}
	if err != nil && pool.StatusCode != 404 {
		log.Warningln("echo")
		log.WithError(err).Panicf("Failed to get pool %v", p.batchConfig.PoolID)
	}

	if err != nil && pool.State == batch.PoolStateActive {
		log.Println("Pool active and running...")
	}

	if pool.Response.StatusCode == 404 {
		// Todo: Fixup pool create currently return error stating SKU not supported
		toCreate := batch.PoolAddParameter{
			ID: &p.batchConfig.PoolID,
			VirtualMachineConfiguration: &batch.VirtualMachineConfiguration{
				ImageReference: &batch.ImageReference{
					Publisher: to.StringPtr("Canonical"),
					Sku:       to.StringPtr("16.04-LTS"),
					Offer:     to.StringPtr("UbuntuServer"),
					Version:   to.StringPtr("latest"),
				},
				NodeAgentSKUID: to.StringPtr("batch.node.ubuntu 16.04"),
			},
			MaxTasksPerNode:      to.Int32Ptr(1),
			TargetDedicatedNodes: to.Int32Ptr(1),
			StartTask: &batch.StartTask{
				ResourceFiles: &[]batch.ResourceFile{
					{
						BlobSource: to.StringPtr("https://raw.githubusercontent.com/Azure/batch-shipyard/f0c9656ca2ccab1a6314f617ff13ea686056f51b/contrib/packer/ubuntu-16.04/bootstrap.sh"),
						FilePath:   to.StringPtr("bootstrap.sh"),
						FileMode:   to.StringPtr("777"),
					},
				},
				CommandLine:    to.StringPtr("bash -f /mnt/batch/tasks/startup/wd/bootstrap.sh 17.12.0~ce-0~ubuntu NVIDIA-Linux-x86_64-384.111.run"),
				WaitForSuccess: to.BoolPtr(true),
				UserIdentity: &batch.UserIdentity{
					AutoUser: &batch.AutoUserSpecification{
						ElevationLevel: batch.Admin,
						Scope:          batch.Pool,
					},
				},
			},
			VMSize: to.StringPtr("standard_a1"),
		}
		poolCreate, err := poolClient.Add(p.ctx, toCreate, nil, nil, nil, nil)

		if err != nil {
			panic(err)
		}

		if poolCreate.StatusCode != 201 {
			panic(poolCreate)
		}

		log.Println("Pool Created")

	}

	for {
		pool, _ := poolClient.Get(p.ctx, p.batchConfig.PoolID, "*", "", nil, nil, nil, nil, "", "", nil, nil)

		if pool.State != "" && pool.State == batch.PoolStateActive {
			log.Println("Created pool... State is Active!")
			break
		} else {
			log.Println("Pool not created yet... sleeping")
			log.Println(pool)
			time.Sleep(time.Second * 20)
		}
	}
}

func createOrGetJob(p *AzureBatch, auth autorest.Authorizer) {
	jobClient := batch.NewJobClientWithBaseURI(getBatchBaseURL(p.batchConfig))
	jobClient.Authorizer = auth
	p.jobClient = &jobClient
	// check if job exists already
	currentJob, err := jobClient.Get(p.ctx, p.dispatcherName, "", "", nil, nil, nil, nil, "", "", nil, nil)

	if err == nil && currentJob.State == batch.JobStateActive {
		log.Println("Wrapper job already exists...")

	} else if currentJob.Response.StatusCode == 404 {

		log.Println("Wrapper job missing... creating...")
		wrapperJob := batch.JobAddParameter{
			ID: &p.dispatcherName,
			PoolInfo: &batch.PoolInformation{
				PoolID: &p.batchConfig.PoolID,
			},
		}

		res, err := jobClient.Add(p.ctx, wrapperJob, nil, nil, nil, nil)

		if err != nil {
			panic(err)
		}

		if res.StatusCode == http.StatusCreated {
			log.WithField("response", res).Info("Job created")
		}

		p.jobClient = &jobClient
	} else if currentJob.State == batch.JobStateDeleting {
		log.Info("Job is being deleted... Waiting then will retry")
		time.Sleep(time.Minute)
		createOrGetJob(p, auth)
	} else {
		log.Info(currentJob)
		panic(err)
	}
}

func getBatchBaseURL(config *types.AzureBatchConfig) string {
	return fmt.Sprintf("https://%s.%s.batch.azure.com", config.BatchAccountName, config.BatchAccountLocation)
}
