package deepstack

import (
	"fmt"
)

type DeepStackError struct {
	Message    string
	StackTrace string
	Context    map[string]any
}

func (d *DeepStackError) Error() string {
	var result = d.Message
	for k, v := range d.Context {
		result += fmt.Sprintf(" %s=%v", k, v)
	}
	result += "\nstack trace:\n" + d.StackTrace
	return result
}
