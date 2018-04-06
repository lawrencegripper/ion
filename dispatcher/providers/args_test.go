package providers

import (
	"testing"
)

func TestGetEnvVaraibles(t *testing.T) {
	envs, err := getModuleEnvironmentVars("./testdata/example.env")
	if err != nil {
		t.Error(err)
	}
	for k, v := range envs {
		t.Logf("Found key: %s and value: %s", k, v)
	}
	if len(envs) != 3 {
		t.Error("Expected to see 3 environment variables in file")
	}
}

func TestGetEnvVaraibles_InvalidFile(t *testing.T) {
	_, err := getModuleEnvironmentVars("./testdata/invalid.env")
	if err == nil {
		t.Error(err)
	}
	t.Log(err)
}

func TestGetEnvVaraibles_MissingFile(t *testing.T) {
	_, err := getModuleEnvironmentVars("./testdata/doesntexist.env")
	if err == nil {
		t.Error(err)
	}
	t.Log(err)
}
