package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestArgumentParsing(t *testing.T) {
    expectedFields := "1,3,5"
    arguments := parseArguments([]string{"-f", expectedFields})
    assert.Equal(t, arguments["fields"], expectedFields)
}
