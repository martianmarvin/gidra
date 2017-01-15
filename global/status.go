package global

// Status Codes
type Status int

const (
	StatusSuccess Status = iota
	StatusFail
	StatusAbort
	StatusSkip
)

// Boolean status helpers to be used in user templates
func (s Status) Success() bool {
	return s == StatusSuccess
}

func (s Status) Fail() bool {
	return s == StatusFail || s == StatusAbort
}

func (s Status) Skip() bool {
	return s == StatusSkip
}
