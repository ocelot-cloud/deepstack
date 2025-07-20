package utils

import (
	"fmt"
)

type DeepStackError struct {
	ErrorMessage string
	ErrorStack   string
	Context      map[string]any
}

func (d *DeepStackError) Error() string {
	var result = d.ErrorMessage
	for k, v := range d.Context {
		result += fmt.Sprintf(" %s=%v", k, v)
	}
	result += "\nstack trace:\n" + d.ErrorStack
	return result
}
