package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
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
