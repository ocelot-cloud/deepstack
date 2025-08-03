package deepstack

type DeepStackError struct {
	Message    string
	StackTrace string
	Context    map[string]any
}

func (d *DeepStackError) Error() string {
	return d.Message
}
