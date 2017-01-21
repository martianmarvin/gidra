package config

//Flags for task Config tags

type Flag byte

const (
	FieldSkip Flag = 1 << iota

	FieldRequired

	FieldOmitEmpty

	// Flags determining when condition should be checked
	// Condition runs before task execution
	CondBefore

	// Condition runs after task execution
	CondAfter

	// Condition only runs once during the entire sequence
	CondOnce
)

func (f *Flag) IsSet(flag Flag) bool {
	return *f&flag != 0
}

func (f *Flag) Set(flag Flag) Flag {
	*f |= flag
	return *f
}

func (f *Flag) UnSet(flag Flag) Flag {
	*f &^= flag
	return *f
}
