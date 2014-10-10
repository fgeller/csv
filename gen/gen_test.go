package main

import "fmt"
import "strings"
import "testing"
import "reflect"

// TODO share
func assert(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Error(
			"Expected", fmt.Sprintf("[%v]", expected),
			"\n",
			"Actual", fmt.Sprintf("[%v]", actual))
	}
}

func TestRandomLineGenerator(t *testing.T) {
	line := randomLine(&parameters{fields: 3})
	assert(t, 2, strings.Count(line, ","))
}
