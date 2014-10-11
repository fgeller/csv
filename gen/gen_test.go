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
	for count := 0; count < 10; count += 1 {
		line := randomLine(&parameters{fields: 3, minWordLength: 1, maxWordLength: 3})
		assert(t, true, 2 <= strings.Count(line, ","))
		assert(t, true, len(line) > 2 && len(line) < 12)
	}
}

func TestRandomFileGenerator(t *testing.T) {
	for count := 0; count < 10; count += 1 {
		file := randomFile(&parameters{fields: 3, minWordLength: 1, maxWordLength: 3, lineCount: 3})
		fmt.Println("Got random file:", file)
		lines := strings.Split(file, "\n")
		assert(t, 3, len(lines))
		for _, line := range lines {
			assert(t, true, 2 <= strings.Count(line, ","))
			assert(t, true, len(line) > 2 && len(line) < 12)
		}
	}
}
