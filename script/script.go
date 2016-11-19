package script

//Script is the runner that contains a series of tasks
type Script struct {
	status int

	//the id of the current iteration
	i int

	//How many times the script should loop total
	Loop int

	//Sequence represents the main task sequence in this script
	Sequence *Sequence

	//BeforeSequence is the sequence to run before running the main one
	BeforeSequence *Sequence
	//AfterSequence is the sequence to run before running the main one
	AfterSequence *Sequence
}

//TODO Load parsed config YAML
// func NewScript(cfg *ScriptConfig) *Script {
// 	return &Script{}
// }

//Finished indicates whether the script completed the current iteration or
//if there are still more tasks remaining
func (s *Script) Finished() bool {
	return s.i > s.Loop
}
