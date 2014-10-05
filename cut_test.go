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

func TestReadingFileToStdOut(t *testing.T) {
    arguments := parseArguments([]string{"-f", "1"})
    fileName := "sample.csv"
    input, _ := os.Open(fileName)
    defer input.Close()
    output := bytes.NewBuffer(nil)
    cut(input, output, arguments)

    assert.Equal(t, "first name,last name,favorite pet", output.String())
}
