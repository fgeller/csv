package main

import "testing"
import "os"
import "bytes"
import "fmt"
import "reflect"
import "strings"

func equal(t *testing.T, expected interface{}, actual interface{}) {

	if !reflect.DeepEqual(expected, actual) {
		t.Error(
			"Expected", fmt.Sprintf("[%v]", expected),
			"\n",
			"Actual", fmt.Sprintf("[%v]", actual))
	}
}

func assert(t *testing.T, assertion interface{}) {
	equal(t, true, assertion)
}

func TestArgumentParsingFailures(t *testing.T) {
	_, msg := parseArguments([]string{"-z"})
	equal(t, "Invalid argument -z", msg)

	_, msg = parseArguments([]string{"idontexist"})
	equal(t, "open idontexist: no such file or directory", msg)
}

func TestArgumentParsingDelimiter(t *testing.T) {
	variations := [][]string{
		[]string{"-d;"},
		[]string{"-d", ";"},
		[]string{"--delimiter", ";"},
		[]string{"--delimiter=;"},
	}

	for _, variation := range variations {
		parameters, messages := parseArguments(variation)
		equal(t, "", messages)
		equal(t, ";", parameters.inputDelimiter)
	}
}

func TestArgumentParsingCSVMode(t *testing.T) {
	variations := [][]string{
		[]string{"-e1-2"},
		[]string{"-e", "1-2"},
	}

	for _, variation := range variations {
		parameters, messages := parseArguments(variation)
		equal(t, "", messages)
		equal(t, []Range{Range{start: 1, end: 2}}, parameters.ranges)
	}
}

func TestArgumentParsingComplement(t *testing.T) {
	parameters, messages := parseArguments([]string{"--complement"})
	equal(t, "", messages)
	assert(t, parameters.complement == true)
}

func TestArgumentParsingOutputDelimiter(t *testing.T) {
	variations := [][]string{
		[]string{"--output-delimiter=|"},
		[]string{"--output-delimiter", "|"},
	}

	for _, variation := range variations {
		parameters, messages := parseArguments(variation)
		equal(t, "", messages)
		equal(t, "|", parameters.outputDelimiter)
	}
}

func TestArgumentParsingHelp(t *testing.T) {
	parameters, messages := parseArguments([]string{"--help"})
	equal(t, "", messages)
	assert(t, parameters.printUsage)
}

func TestArgumentParsingVersion(t *testing.T) {
	parameters, messages := parseArguments([]string{"--version"})
	equal(t, "", messages)
	assert(t, parameters.printVersion)
}

func TestColumnsArgumentParsing(t *testing.T) {

	arguments, _ := parseArguments([]string{"-e", "1-3"})
	equal(t, []Range{NewRange(1, 3)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-e", "1-"})
	equal(t, []Range{NewRange(1, 0)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-e", "1"})
	equal(t, []Range{NewRange(1, 1)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-e", "-23"})
	equal(t, []Range{NewRange(0, 23)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-e", "1-3,5"})
	equal(t, []Range{NewRange(1, 3), NewRange(5, 5)}, arguments.ranges)

	arguments, _ = parseArguments([]string{"-e", "1-3,-5,23,42-"})
	equal(t, []Range{NewRange(1, 3), NewRange(0, 5), NewRange(23, 23), NewRange(42, 0)}, arguments.ranges)
}

func TestDelimiterArgumentParsing(t *testing.T) {
	arguments, _ := parseArguments([]string{"-d", ","})
	equal(t, ",", arguments.inputDelimiter)

	arguments, _ = parseArguments([]string{"-d,"})
	equal(t, ",", arguments.inputDelimiter)

	arguments, _ = parseArguments([]string{})
	equal(t, ",", arguments.inputDelimiter)
}

func TestFileNameArgumentParsing(t *testing.T) {
	arguments, _ := parseArguments([]string{"sample.csv"})
	equal(t, "sample.csv", arguments.input[0].Name())

	arguments, _ = parseArguments([]string{})
	equal(t, []*os.File{os.Stdin}, arguments.input)

	arguments, _ = parseArguments([]string{"-"})
	equal(t, []*os.File{os.Stdin}, arguments.input)
}

var fullFile = `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch`

var cutTests = []struct {
	parameters []string
	input      string
	expected   string
}{
	{ // inversing range
		parameters: []string{"-d,", "-e-2", "--complement", "--line-end=LF"},
		input:      fullFile,
		expected: `favorite pet
moose
monarch
`,
	},
	{ // cutting first and second column via range
		parameters: []string{"-d,", "-e1-2", "--line-end=LF"},
		input:      fullFile,
		expected: `first name,last name
hans,hansen
peter,petersen
`,
	},
	{ // cutting all via a range
		parameters: []string{"-d,", "-e1-", "--line-end=LF"},
		input:      fullFile,
		expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
	},
	{ // cutting all via a range
		parameters: []string{"-d,", "-e-3", "--line-end=LF"},
		input:      fullFile,
		expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
	},
	{ // cutting all via a range
		parameters: []string{"-d,", "-e1-3,3", "--line-end=LF"},
		input:      fullFile,
		expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
	},
	{ // cutting fields with multi-byte delimiter
		parameters: []string{"-d€", "-e2", "--line-end=LF"},
		input: "first name€last name€favorite pet\x0a" +
			"hans€hansen€moose\x0a" +
			"peter€petersen€monarch\x0a",
		expected: `last name
hansen
petersen
`,
	},
	{ // cutting fields separated by spaces
		parameters: []string{"-d ", "-e2", "--line-end=LF"},
		input: "first second third\x0a" +
			"a b c\x0a" +
			"d e f\x0a",
		expected: "second\x0ab\x0ae\x0a",
	},
	{ // cutting fields separated by quotes
		parameters: []string{"-d'", "-e2", "--line-end=LF"},
		input: "first'second'third\x0a" +
			"a'b'c\x0a" +
			"d'e'f\x0a",
		expected: "second\x0ab\x0ae\x0a",
	},
	{ // cutting csv values with LF rather than CRLF line ending
		parameters: []string{"-e2-", "--line-end=LF"},
		input: "first a,last b,favorite pet\x0a" +
			"hans,hansen,moose\x0a" +
			"peter,petersen,monarch\x0a",
		expected: "last b,favorite pet\x0a" +
			"hansen,moose\x0a" +
			"petersen,monarch\x0a",
	},
	{ // cutting csv values with CRLF explicitly
		parameters: []string{"-e2-", "--line-end=CRLF"},
		input: "first a,last b,favorite pet\x0d\x0a" +
			"hans,hansen,moose\x0d\x0a" +
			"peter,petersen,monarch\x0d\x0a",
		expected: "last b,favorite pet\x0d\x0a" +
			"hansen,moose\x0d\x0a" +
			"petersen,monarch\x0d\x0a",
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
	{ // cutting csv values with custom input delimiters
		parameters: []string{"-e2-", "-d;"},
		input: "first a;last a;favorite pet\x0d\x0a" +
			"hans;hansen;moose\x0d\x0a" +
			"peter;petersen;monarch\x0d\x0a",
		expected: "last a;favorite pet\x0d\x0a" +
			"hansen;moose\x0d\x0a" +
			"petersen;monarch\x0d\x0a",
	},
	{ // cutting csv values with custom multi-byte input delimiters
		parameters: []string{"-e2-", "-d€", "--output-delimiter=;"},
		input: "first a€last a€favorite pet\x0d\x0a" +
			"hans€hansen€moose\x0d\x0a" +
			"peter€petersen€monarch\x0d\x0a",
		expected: "last a;favorite pet\x0d\x0a" +
			"hansen;moose\x0d\x0a" +
			"petersen;monarch\x0d\x0a",
	},
	{ // cutting csv values with custom input and output delimiters
		parameters: []string{"-e2-", "-d;", "--output-delimiter=|"},
		input: "first a;last a;favorite pet\x0d\x0a" +
			"hans;hansen;moose\x0d\x0a" +
			"peter;petersen;monarch\x0d\x0a",
		expected: "last a|favorite pet\x0d\x0a" +
			"hansen|moose\x0d\x0a" +
			"petersen|monarch\x0d\x0a",
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
}

func TestCutFile(t *testing.T) {
	for _, data := range cutTests {
		parameters, _ := parseArguments(data.parameters)
		input := strings.NewReader(data.input)
		output := bytes.NewBuffer(nil)

		cutFile(input, output, parameters)

		equal(t, data.expected, output.String())
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

	cut([]string{fileName, "--line-end=LF"}, output)

	equal(t, string(contents), output.String())
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

	cut([]string{fileName, fileName, "--line-end=LF"}, output)

	equal(t, fmt.Sprint(string(contents), string(contents)), output.String())
}

func TestPrintingUsageInformation(t *testing.T) {
	output := bytes.NewBuffer(nil)
	cut([]string{"--help"}, output)

	equal(t, true, strings.HasPrefix(output.String(), "Usage: "))
}
