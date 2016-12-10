package config

//Flags for task Config tags

type Flag byte

const (
	FieldSkip Flag = 1 << iota

	FieldRequired

	FieldOmitEmpty
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
