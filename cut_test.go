package main

import "testing"
import "os"
import "bytes"
import "fmt"
import "reflect"

func assert(t *testing.T, expected interface{}, actual interface{}) {
    if !reflect.DeepEqual(expected, actual) {
        t.Error(
            "Expected", fmt.Sprintf("[%v]", expected),
            "\n",
            "Actual", fmt.Sprintf("[%v]", actual))
    }
}

func TestFieldsArgumentParsing(t *testing.T) {
    expectedFields := "1,3,5"

    arguments, _ := parseArguments([]string{fmt.Sprint("-f", expectedFields)})
    assert(t, []int64{1, 3, 5}, arguments.fields)

    arguments, _ = parseArguments([]string{"-f", expectedFields})
    assert(t, []int64{1, 3, 5}, arguments.fields)

    arguments, _ = parseArguments([]string{})
    assert(t, []int64{}, arguments.fields)
}

func TestDelimiterArgumentParsing(t *testing.T) {
    arguments, _ := parseArguments([]string{"-d", ","})
    assert(t, ",", arguments.delimiter)

    arguments, _ = parseArguments([]string{"-d,"})
    assert(t, ",", arguments.delimiter)

    arguments, _ = parseArguments([]string{})
    assert(t, ",", arguments.delimiter)
}

func TestFileNameArgumentParsing(t *testing.T) {
    arguments, _ := parseArguments([]string{"sample.csv"})
    assert(t, "sample.csv", arguments.files[0].Name())

    arguments, _ = parseArguments([]string{})
    assert(t, []*os.File{}, arguments.files)
}

var cutTests = []struct {
    selectedFields []int64
    delimiter      string
    expected       string
}{
    { // full file when no fields
        selectedFields: []int64{},
        delimiter:      ",",
        expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
    },
    { // full file when no delimiter
        selectedFields: []int64{1},
        delimiter:      "\t",
        expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
    },
    { // cutting first column
        selectedFields: []int64{1},
        delimiter:      ",",
        expected: `first name
hans
peter
`,
    },
    { // cutting second column
        selectedFields: []int64{2},
        delimiter:      ",",
        expected: `last name
hansen
petersen
`,
    },
    { // cutting third column
        selectedFields: []int64{3},
        delimiter:      ",",
        expected: `favorite pet
moose
monarch
`,
    },
    { // cutting first and third column
        selectedFields: []int64{1, 3},
        delimiter:      ",",
        expected: `first name,favorite pet
hans,moose
peter,monarch
`,
    },
}

func TestCutFile(t *testing.T) {
    fileName := "sample.csv"

    for _, data := range cutTests {
        input, _ := os.Open(fileName)
        defer input.Close()
        output := bytes.NewBuffer(nil)
        cutFile(input, output, data.delimiter, data.selectedFields)

        assert(t, output.String(), data.expected)
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
