package script

//FieldError is a custom error for missing fields in the config
type FieldError struct {
	Name string
}

func (e FieldError) Error() string {
	return "Config is missing required field: " + e.Name
}
