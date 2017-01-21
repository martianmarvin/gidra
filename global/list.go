package global

import (
	"container/ring"
	"fmt"
)

// Lists are simple lists of values that are easy for a user to manipulate from
// within templates
// Lists are not concurrency-safe, so should not be reused across sequences
// List operate as a ring, constantly looping through values, preventing EOF
// errors
// Elements cannot be added to the list once it has been initialized - in that
// case, a new List should be initialized
type List struct {
	index int
	r     *ring.Ring
}

// NewList initializes a new list
func NewList(vals []interface{}) *List {
	l := &List{r: ring.New(len(vals))}
	for i := 0; i < l.r.Len(); i++ {
		l.r.Value = vals[i]
		l.r = l.r.Next()
	}
	return l
}

// Next advances the list to the next value
func (l *List) Next() interface{} {
	if l.index >= l.Len() {
		l.index = 0
	} else {
		l.index++
	}
	l.r = l.r.Next()
	return l.r.Value
}

// Returns the current value in the list
func (l *List) Value() interface{} {
	return l.r.Value
}

func (l *List) Len() int {
	return l.r.Len()
}

// Values returns all values in the list
func (l *List) Values() []interface{} {
	var vs []interface{}
	l.r.Do(func(v interface{}) {
		vs = append(vs, v)
	})
	return vs
}

// Removes the top value from the list
func (l *List) Pop() interface{} {
	if l.r.Len() == 0 {
		return nil
	}
	l.r = l.r.Prev()
	return l.r.Unlink(1).Value
}

// String returns the CURRENT value of the list as a string
func (l *List) String() string {
	return fmt.Sprint(l.Value())
}
