package tools

import "encoding/json"

// PrettyPrintStruct give a indented JSON representation of a struct
func PrettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}
