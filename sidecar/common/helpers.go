package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
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

//StripBlobStore removes details specific to the metadata store from any metadata
func StripBlobStore(docs []MetaDoc) ([]MetaDoc, error) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(docs))

	defer waitGroup.Wait()

	rx, err := regexp.Compile(`^(?:https?:\/\/)?(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)/[a-zA-Z0-9]+/`)
	if err != nil {
		return nil, fmt.Errorf("error thrown compiling regex: %+v", err)
	}

	strip := make(chan MetaDoc)
	strippedDocs := make([]MetaDoc, 0)
	for _, doc := range docs {
		go func(doc MetaDoc) {
			defer waitGroup.Done()
			strippedMeta := map[string]string{}
			for k, v := range doc.Metadata {
				match := rx.FindString(v)
				if match == "" {
					strippedMeta[k] = v
				} else {
					strippedMeta[k] = strings.Replace(v, match, "", 1)
				}
			}
			doc.Metadata = strippedMeta
			strip <- doc
		}(doc)
	}

	for i := 0; i < len(docs); i++ {
		doc := <-strip
		strippedDocs = append(strippedDocs, doc)
	}

	return strippedDocs, nil
}

//NormalizeResourcePath transforms a resource path into an expected format
func NormalizeResourcePath(resPath string) (string, error) {
	if resPath[0] == '/' {
		resPath = resPath[1:]
	}
	segs := strings.Split(resPath, "/")
	if len(segs) < 2 {
		return "", fmt.Errorf("%s is not a valid resource path", resPath)
	}
	resPath = strings.Replace(resPath, "//", "/", -1)
	return resPath, nil
}
