package deepstack

import "fmt"

type DeepStackError struct {
	Message    string
	StackTrace string
	Context    map[string]any
}

func (d *DeepStackError) Error() string {
	return fmt.Sprintf("message: %s; context: %s; stack: %s", d.Message, d.Context, d.StackTrace)
}
