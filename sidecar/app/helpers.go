package app

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/twinj/uuid"
)

//CompareHash compares a secret string against a hash
func CompareHash(secret, secretHash string) error {
	if secret == "" {
		return fmt.Errorf("secret header was not found")
	}
	if Hash(secret) != secretHash {
		return fmt.Errorf("secret did not match")
	}
	return nil
}

//Hash returns a MD5 hash of the provided string
// nolint: errcheck
func Hash(s string) string {
	hasher := md5.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

//MustNotBeEmpty panics if any of the strings provided are empty
func MustNotBeEmpty(strs ...string) {
	for _, s := range strs {
		if s == "" {
			panic("required string is empty")
		}
	}
}

//MustNotBeNil panics if any of the objects provided are nil
func MustNotBeNil(objs ...interface{}) {
	for _, o := range objs {
		if o == nil {
			panic("required obj is nil")
		}
	}
}

func ClearDir(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed removing directory path '%s' with error: '%+v'", dirPath, err)
	}
	err = os.MkdirAll(dirPath, 0777)
	if err != nil {
		return fmt.Errorf("failed creating directory path '%s' with error: '%+v'", dirPath, err)
	}
	return nil
}

func RemoveFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to remove file at path '%s' with error: '%+v'", filePath, err)
	}
	return nil
}

func NewExecutionID(moduleName string) string {
	guid := fmt.Sprintf("%v", uuid.NewV4())
	return moduleName + "_" + guid
}
