package providers

import (
	apiv1 "k8s.io/api/core/v1"
	"strings"
	"testing"
)

func TestVolumeGeneration(t *testing.T) {

	containers := []apiv1.Container{
		{
			Name:  "sidecar",
			Image: "doesntmatter",
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ionvolume",
					MountPath: "/ion",
				},
			},
		},
		{
			Name:            "worker",
			Image:           "doesntmatter",
			ImagePullPolicy: apiv1.PullAlways,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ionvolume",
					MountPath: "/ion",
				},
			},
		},
	}

	// Todo: Pull this out into a standalone package once stabilized
	podCommand, err := getPodCommand(batchPodComponents{
		Containers: containers,
		PodName:    mockMessageID,
		TaskID:     mockMessageID,
		Volumes: []apiv1.Volume{
			{
				Name: "ionvolume",
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			},
		},
	})

	if err != nil {
		t.Error(err)
	}

	// Todo: Very basic smoke test that shared volume path is present in batch command
	if !strings.Contains(podCommand, "/ion") {
		t.Log(podCommand)
		t.Error("Missing shared volume")
	}

	// Todo: Very basic smoke test that shared volume path is present in batch command
	if !strings.Contains(podCommand, "docker volume create examplemessageID_ionvolume") {
		t.Log(podCommand)
		t.Error("Missing shared volume")
	}
}

func TestPod2DockerGeneratesValidOutputEncoding(t *testing.T) {
	containers := []apiv1.Container{
		{
			Name:  "sidecar",
			Image: "barry",
			Args:  []string{"encoding"},
		},
		{
			Name:            "worker",
			Image:           "marge",
			ImagePullPolicy: apiv1.PullAlways,
		},
	}

	// Todo: Pull this out into a standalone package once stabilized
	podCommand, err := getPodCommand(batchPodComponents{
		Containers: containers,
		PodName:    mockMessageID,
		TaskID:     mockMessageID,
		Volumes:    nil,
	})

	if err != nil {
		t.Error(err)
	}

	t.Log(podCommand)
	if strings.Contains(podCommand, "&lt;") {
		t.Error("output contains incorrect encoding")
	}
}
