package main

import "fmt"
import "strings"
import "testing"
import "reflect"
import "bytes"

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

		lines := strings.Split(file, "\n")
		assert(t, 3, len(lines))
		for _, line := range lines {
			assert(t, true, 2 <= strings.Count(line, ","))
			assert(t, true, len(line) > 2 && len(line) < 12)
		}
	}
}

var genTestData = []struct {
	arguments     []string
	fields        int
	minWordLength int
	maxWordLength int
	lineCount     int
}{
	{
		arguments:     []string{"-l2", "-f3", "-cmax10", "-cmin1"},
		fields:        3,
		minWordLength: 1,
		maxWordLength: 10,
		lineCount:     2,
	},
}

func TestGen(t *testing.T) {

	for _, data := range genTestData {
		output := bytes.NewBuffer(nil)
		gen(data.arguments, output)

		generated := output.String()

		lines := strings.Split(generated, "\n")
		assert(t, data.lineCount+1, len(lines))

		for _, line := range lines[:data.lineCount] {
			assert(t, true, data.fields-1 <= strings.Count(line, ","))
			assert(t, true, len(line) > (data.fields-1+data.fields*data.minWordLength))
			assert(t, true, len(line) < (data.fields*data.maxWordLength+data.fields))
		}
	}
}
