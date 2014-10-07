package main

import "testing"
import "github.com/stretchr/testify/assert"
import "os"
import "bytes"

func TestArgumentParsing(t *testing.T) {
    expectedFields := "1,3,5"
    arguments := parseArguments([]string{"-f", expectedFields})
    assert.Equal(t, expectedFields, arguments["fields"])
}

func TestCopyingEntireFile(t *testing.T) {
    fileName := "sample.csv"
    expected := `first name,last name,favorite pet
hans,hansen,moose
peter,petersen,monarch
`
    input, _ := os.Open(fileName)
    defer input.Close()
    output := bytes.NewBuffer(nil)
    cut(input, output, "\t", []int64{1})

    assert.Equal(t, string(expected), output.String())
}

func TestCuttingFirstColumn(t *testing.T) {
    fileName := "sample.csv"
    expected := `first name
hans
peter
`
    input, _ := os.Open(fileName)
    defer input.Close()
    output := bytes.NewBuffer(nil)
    cut(input, output, ",", []int64{1})

    assert.Equal(t, string(expected), output.String())
}

func TestCuttingSecondColumn(t *testing.T) {
    fileName := "sample.csv"
    expected := `last name
hansen
petersen
`
    input, _ := os.Open(fileName)
    defer input.Close()
    output := bytes.NewBuffer(nil)
    cut(input, output, ",", []int64{2})

    assert.Equal(t, string(expected), output.String())
}

func TestCuttingThirdColumn(t *testing.T) {
    fileName := "sample.csv"
    expected := `favorite pet
moose
monarch
`
    input, _ := os.Open(fileName)
    defer input.Close()
    output := bytes.NewBuffer(nil)
    cut(input, output, ",", []int64{3})

    assert.Equal(t, string(expected), output.String())
}

func TestCuttingFirstAndThirdColumn(t *testing.T) {
    fileName := "sample.csv"
    expected := `first name,favorite pet
hans,moose
peter,monarch
`
    input, _ := os.Open(fileName)
    defer input.Close()
    output := bytes.NewBuffer(nil)
    cut(input, output, ",", []int64{1, 3})

    assert.Equal(t, string(expected), output.String())
}
