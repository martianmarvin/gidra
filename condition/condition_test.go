package condition

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var trueTmpl = []string{`{{ eq 1 1 }}`, `{{ eq "true" "true" }}`, `{{ eq .Name "user" }}`}
var falseTmpl = []string{`{{ eq 1 2 }}`, `{{ eq "true" "false" }}`, `{{ eq .Name "notUser" }}`}

var cvars = map[string]interface{}{
	"Name": "user",
}

// Test a condition, and assert that it returns errT when MET, and errF when
// NOT MET
func testCondition(t *testing.T, cond Condition, errT, errF error) {
	var err error
	assert := assert.New(t)

	for _, tmpl := range trueTmpl {
		err = cond.Parse(tmpl)
		assert.NoError(err)

		err = cond.Check(cvars)
		if errT == nil {
			assert.NoError(err)
		} else {
			assert.EqualError(err, errT.Error())
		}
	}

	for _, tmpl := range falseTmpl {
		err = cond.Parse(tmpl)
		assert.NoError(err)

		err = cond.Check(cvars)
		if errF == nil {
			assert.NoError(err)
		} else {
			assert.EqualError(err, errF.Error())
		}
	}
}

func TestSuccess(t *testing.T) {
	testCondition(t, NewSuccess(), nil, ErrFail)
}

func TestOnly(t *testing.T) {
	testCondition(t, NewOnly(), nil, ErrSkip)
}

func TestSkip(t *testing.T) {
	testCondition(t, NewSkip(), ErrSkip, nil)
}

func TestAbort(t *testing.T) {
	testCondition(t, NewAbort(), ErrAbort, nil)
}

func TestRetry(t *testing.T) {
	cond := NewRetry(len(falseTmpl))
	testCondition(t, cond, ErrRetry, nil)
	testCondition(t, cond, ErrFail, nil)
}

func TestFail(t *testing.T) {
	testCondition(t, NewFail(), ErrFail, nil)
}

// TODO Test that callbacks fire in correct order
