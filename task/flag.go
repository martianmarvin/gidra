package task

//Flags for task Config tags

type ConfigFlag byte

const (
	FieldSkip ConfigFlag = 1 << iota

	FieldRequired

	FieldOmitEmpty
)

func (f *ConfigFlag) IsSet(flag ConfigFlag) bool {
	return *f&flag != 0
}

func (f *ConfigFlag) Set(flag ConfigFlag) ConfigFlag {
	*f |= flag
	return *f
}

func (f *ConfigFlag) UnSet(flag ConfigFlag) ConfigFlag {
	*f &^= flag
	return *f
}
