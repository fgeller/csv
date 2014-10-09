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
    assert(t, ",", arguments.delimiter)

    arguments, _ = parseArguments([]string{"-d,"})
    assert(t, ",", arguments.delimiter)

    arguments, _ = parseArguments([]string{})
    assert(t, ",", arguments.delimiter)
}

func TestFileNameArgumentParsing(t *testing.T) {
    arguments, _ := parseArguments([]string{"sample.csv"})
    assert(t, "sample.csv", arguments.input[0].Name())

    arguments, _ = parseArguments([]string{})
    assert(t, []*os.File{os.Stdin}, arguments.input)

    arguments, _ = parseArguments([]string{"-"})
    assert(t, []*os.File{os.Stdin}, arguments.input)
}

var cutTests = []struct {
    ranges    []Range
    delimiter string
    expected  string
}{
    { // full file when no fields
        ranges:    []Range{},
        delimiter: ",",
        expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
    },
    { // full file when no delimiter
        ranges:    []Range{NewRange(1, 1)},
        delimiter: `\t`,
        expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
    },
    { // cutting first column
        ranges:    []Range{NewRange(1, 1)},
        delimiter: ",",
        expected: `first name
hans
peter
`,
    },
    { // cutting second column
        ranges:    []Range{NewRange(2, 2)},
        delimiter: ",",
        expected: `last name
hansen
petersen
`,
    },
    { // cutting third column
        ranges:    []Range{NewRange(3, 3)},
        delimiter: ",",
        expected: `favorite pet
moose
monarch
`,
    },
    { // cutting first and third column
        ranges:    []Range{NewRange(1, 1), NewRange(3, 3)},
        delimiter: ",",
        expected: `first name,favorite pet
hans,moose
peter,monarch
`,
    },
    { // cutting first and second column via range
        ranges:    []Range{NewRange(1, 2)},
        delimiter: ",",
        expected: `first name,last name
hans,hansen
peter,petersen
`,
    },
    { // cutting all via a range
        ranges:    []Range{NewRange(1, 0)},
        delimiter: ",",
        expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
    },
    { // cutting all via a range
        ranges:    []Range{NewRange(0, 3)},
        delimiter: ",",
        expected: `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`,
    },
}

func TestCutFile(t *testing.T) {
    fileName := "sample.csv"

    for _, data := range cutTests {
        input, _ := os.Open(fileName)
        defer input.Close()
        output := bytes.NewBuffer(nil)

        cutFile(input, output, data.delimiter, data.ranges)

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
