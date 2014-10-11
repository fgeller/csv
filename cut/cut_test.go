package main

import "testing"
import "os"
import "bytes"
import "fmt"
import "reflect"
import "strings"

func assert(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Error(
			"Expected", fmt.Sprintf("[%v]", expected),
			"\n",
			"Actual", fmt.Sprintf("[%v]", actual))
	}
}

func TestFieldsArgumentParsing(t *testing.T) {
	arguments, _ := parseArguments([]string{fmt.Sprint("-f", "1,3,5")})
	assert(t, []Range{NewRange(1, 1), NewRange(3, 3), NewRange(5, 5)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-f", "1,3,5"})
	assert(t, []Range{NewRange(1, 1), NewRange(3, 3), NewRange(5, 5)}, arguments.ranges)

	arguments, _ = parseArguments([]string{})
	assert(t, []Range{}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-f", "1-3"})
	assert(t, []Range{NewRange(1, 3)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-f", "1-"})
	assert(t, []Range{NewRange(1, 0)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-f", "1"})
	assert(t, []Range{NewRange(1, 1)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-f", "-23"})
	assert(t, []Range{NewRange(0, 23)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-f", "1-3,5"})
	assert(t, []Range{NewRange(1, 3), NewRange(5, 5)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-f", "1-3,-5,23,42-"})
	assert(t, []Range{NewRange(1, 3), NewRange(0, 5), NewRange(23, 23), NewRange(42, 0)}, arguments.ranges)
}

func TestDelimiterArgumentParsing(t *testing.T) {
	arguments, _ := parseArguments([]string{"-d", ","})
	assert(t, ",", arguments.inputDelimiter)

	arguments, _ = parseArguments([]string{"-d,"})
	assert(t, ",", arguments.inputDelimiter)

	arguments, _ = parseArguments([]string{})
	assert(t, "\t", arguments.inputDelimiter)
}

func TestFileNameArgumentParsing(t *testing.T) {
	arguments, _ := parseArguments([]string{"sample.csv"})
	assert(t, "sample.csv", arguments.input[0].Name())

	arguments, _ = parseArguments([]string{})
	assert(t, []*os.File{os.Stdin}, arguments.input)

	arguments, _ = parseArguments([]string{"-"})
	assert(t, []*os.File{os.Stdin}, arguments.input)
}

var fullFile = `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch`

var cutTests = []struct {
	parameters []string
	input      string
	expected   string
}{
	{ // full file when no fields
		parameters: []string{"-d,"},
		input:      fullFile,
		expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
	},
	{ // full file when no delimiter XXXX
		parameters: []string{"-dx", "-f1"},
		input:      fullFile,
		expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
	},
	{ // cutting first column
		parameters: []string{"-d,", "-f1"},
		input:      fullFile,
		expected: `first name
hans
peter
`,
	},
	{ // cutting second column
		parameters: []string{"-d,", "-f2"},
		input:      fullFile,
		expected: `last name
hansen
petersen
`,
	},
	{ // cutting third column
		parameters: []string{"-d,", "-f3"},
		input:      fullFile,
		expected: `favorite pet
moose
monarch
`,
	},
	{ // inversing range
		parameters: []string{"-d,", "-f-2", "--complement"},
		input:      fullFile,
		expected: `favorite pet
moose
monarch
`,
	},
	{ // cutting first and third column
		parameters: []string{"-d,", "-f1,3"},
		input:      fullFile,
		expected: `first name,favorite pet
hans,moose
peter,monarch
`,
	},
	{ // cutting first and second column via range
		parameters: []string{"-d,", "-f1-2"},
		input:      fullFile,
		expected: `first name,last name
hans,hansen
peter,petersen
`,
	},
	{ // cutting all via a range
		parameters: []string{"-d,", "-f1-"},
		input:      fullFile,
		expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
	},
	{ // cutting all via a range
		parameters: []string{"-d,", "-f-3"},
		input:      fullFile,
		expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
	},
	{ // cutting all via a range
		parameters: []string{"-d,", "-f1-3,3"},
		input:      fullFile,
		expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
	},
	{ // cutting csv values
		parameters: []string{"-e2-"},
		input: "first a,last a,favorite pet\x0d\x0a" +
			"hans,hansen,moose\x0d\x0a" +
			"peter,petersen,monarch\x0d\x0a",
		expected: "last a,favorite pet\x0d\x0a" +
			"hansen,moose\x0d\x0a" +
			"petersen,monarch\x0d\x0a",
	},
	{ // cutting csv values that are escaped
		parameters: []string{"-e2-3"},
		input: "first name,last name,\"favorite pet\"\x0d\x0a" +
			"\"hans\",hansen,\"moose,goose\"\x0d\x0a" +
			"peter,\"petersen,muellersen\",monarch\x0d\x0a",
		expected: "last name,\"favorite pet\"\x0d\x0a" +
			"hansen,\"moose,goose\"\x0d\x0a" +
			"\"petersen,muellersen\",monarch\x0d\x0a",
	},
	{ // cutting csv values that are escaped and contain new lines
		parameters: []string{"-e2-3"},
		input: "first name,last name,\"\x0d\x0afavorite pet\"\x0d\x0a" +
			"\"hans\",hansen,\"moose,goose\"\x0d\x0a" +
			"peter,\"petersen,muellersen\x0d\x0a\",monarch\x0d\x0a",
		expected: "last name,\"\x0d\x0afavorite pet\"\x0d\x0a" +
			"hansen,\"moose,goose\"\x0d\x0a" +
			"\"petersen,muellersen\x0d\x0a\",monarch\x0d\x0a",
	},
	{ // cutting csv values that are doubly escaped
		parameters: []string{"-e2-3"},
		input: "first name,last name,\"favorite\"\" pet\"\x0d\x0a" +
			"\"hans\",hansen,\"moose,goose\"\x0d\x0a" +
			"peter,\"petersen,\"\"\"\"\"\"\"\"muellersen\",monarch\x0d\x0a",
		expected: "last name,\"favorite\"\" pet\"\x0d\x0a" +
			"hansen,\"moose,goose\"\x0d\x0a" +
			"\"petersen,\"\"\"\"\"\"\"\"muellersen\",monarch\x0d\x0a",
	},
	{ // select bytes
		parameters: []string{"-b-2"},
		input:      `€foo`,
		expected:   "\xe2\x82\x0a",
	},
	{ // select characters / runes
		parameters: []string{"-c-2"},
		input:      `€foo`,
		expected: `€f
`,
	},
	{ // select characters / runes with custom separator
		parameters: []string{"-c-2", "--output-delimiter", "x"},
		input:      `€foo`,
		expected: `€xf
`,
	},
	{ // include lines that don't contain delimiter by default
		parameters: []string{"-d,", "-f2"},
		input: `first name,last name
no delimiter here
same name,and another`,
		expected: `last name
no delimiter here
and another
`,
	},
	{ // include exclude lines without delimiter
		parameters: []string{"-d,", "-f2", "-s"},
		input: `first name,last name
no delimiter here
same name,and another`,
		expected: `last name
and another
`,
	},
	{ // include exclude lines without delimiter
		parameters: []string{"-d,", "--only-delimited", "-f2"},
		input: `first name,last name
no delimiter here
same name,and another`,
		expected: `last name
and another
`,
	},
	{ // include exclude lines without delimiter
		parameters: []string{"--output-delimiter", "x", "-d,", "--only-delimited", "-f1,2"},
		input: `first name,last name
no delimiter here
same name,and another`,
		expected: `first namexlast name
same namexand another
`,
	},
	{ // ignore -n
		parameters: []string{"--output-delimiter", "x", "-n", "-d,", "--only-delimited", "-f1,2"},
		input: `first name,last name
no delimiter here
same name,and another`,
		expected: `first namexlast name
same namexand another
`,
	},
}

func TestCutFile(t *testing.T) {
	for _, data := range cutTests {
		parameters, _ := parseArguments(data.parameters)
		input := strings.NewReader(data.input)
		output := bytes.NewBuffer(nil)

		cutFile(input, output, parameters)

		assert(t, data.expected, output.String())
	}
}

func TestCut(t *testing.T) {
	fileName := "sample.csv"
	contents := `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`

	input, _ := os.Open(fileName)
	defer input.Close()
	output := bytes.NewBuffer(nil)

	cut([]string{fileName}, output)

	assert(t, string(contents), output.String())
}

func TestCuttingMultipleFiles(t *testing.T) {
	fileName := "sample.csv"
	contents := `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`

	input, _ := os.Open(fileName)
	defer input.Close()
	output := bytes.NewBuffer(nil)

	cut([]string{fileName, fileName}, output)

	assert(t, fmt.Sprint(string(contents), string(contents)), output.String())
}
