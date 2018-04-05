package providers

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
import (
	"bytes"
	"fmt"
	"github.com/lawrencegripper/ion/dispatcher/types"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/Azure/go-autorest/autorest"

	"github.com/Azure/azure-sdk-for-go/services/batch/2017-09-01.6.0/batch"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/api/core/v1"
)

func createOrGetPool(p *AzureBatch, auth autorest.Authorizer) {

	poolClient := batch.NewPoolClientWithBaseURI(getBatchBaseURL(p.batchConfig))
	poolClient.Authorizer = auth
	poolClient.RetryAttempts = 0
	//	poolClient.RequestInspector = fixContentTypeInspector()
	p.poolClient = &poolClient
	pool, err := poolClient.Get(p.ctx, p.batchConfig.PoolID, "*", "", nil, nil, nil, nil, "", "", nil, nil)

	// If we observe an error which isn't related to the pool not existing panic.
	// 404 is expected if this is first run.
	if err != nil && pool.StatusCode != 404 {
		panic(err)
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
	} else {
		// unknown case...
		panic(err)
	}
}

func getBatchBaseURL(config *types.AzureBatchConfig) string {
	return fmt.Sprintf("https://%s.%s.batch.azure.com", config.BatchAccountName, config.BatchAccountLocation)
}

func getLaunchCommand(container v1.Container) (cmd string) {
	if len(container.Command) > 0 {
		cmd += strings.Join(container.Command, " ")
	}
	if len(cmd) > 0 {
		cmd += " "
	}
	if len(container.Args) > 0 {
		cmd += strings.Join(container.Args, " ")
	}
	return
}

func getPodCommand(p batchPodComponents) (string, error) {
	template := template.New("run.sh.tmpl").Option("missingkey=error").Funcs(template.FuncMap{
		"getLaunchCommand":     getLaunchCommand,
		"isHostPathVolume":     isHostPathVolume,
		"isEmptyDirVolume":     isEmptyDirVolume,
		"isPullAlways":         isPullAlways,
		"getValidVolumeMounts": getValidVolumeMounts,
	})

	template, err := template.Parse(azureBatchPodTemplate)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	err = template.Execute(&output, p)
	return output.String(), err
}

func isHostPathVolume(v v1.Volume) bool {
	if v.HostPath == nil {
		return false
	}
	return true
}

func isEmptyDirVolume(v v1.Volume) bool {
	if v.EmptyDir == nil {
		return false
	}
	return true
}

func isPullAlways(c v1.Container) bool {
	if c.ImagePullPolicy == v1.PullAlways {
		return true
	}
	return false
}

func getValidVolumeMounts(container v1.Container, volumes []v1.Volume) []v1.VolumeMount {
	volDic := make(map[string]v1.Volume)
	for _, vol := range volumes {
		volDic[vol.Name] = vol
	}
	var mounts []v1.VolumeMount
	for _, mount := range container.VolumeMounts {
		vol, ok := volDic[mount.Name]
		if !ok {
			continue
		}
		if vol.EmptyDir == nil && vol.HostPath == nil {
			continue
		}
		mounts = append(mounts, mount)
	}
	return mounts
}
