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
            "Actual", fmt.Sprintf("[%v]", actual))
    }
}

func TestFieldsArgumentParsing(t *testing.T) {
    expectedFields := "1,3,5"
    arguments := parseArguments([]string{"-f", expectedFields})
    fields := selectedFields(arguments)
    assert(t, []int64{1, 3, 5}, fields)
}

func TestDelimiterArgumentParsing(t *testing.T) {
    arguments := parseArguments([]string{"-d", ","})
    assert(t, ",", delimiter(arguments))
}

var cutTests = []struct {
    selectedFields []int64
    delimiter      string
    expected       string
}{
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

func TestCut(t *testing.T) {
    fileName := "sample.csv"

    for _, data := range cutTests {
        input, _ := os.Open(fileName)
        defer input.Close()
        output := bytes.NewBuffer(nil)
        cut(input, output, data.delimiter, data.selectedFields)

        if output.String() != data.expected {
            t.Error(
                "Expected", fmt.Sprintf("[%v]", data.expected),
                "Actual", fmt.Sprintf("[%v]", output.String()))
        }
    }
}
