package types

import "encoding/json"

// PrettyPrintStruct outputs the json form of the struct
func PrettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}
