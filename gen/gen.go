package main

import "math/rand"
import "time"
import "strings"
import "os"
import "io"
import "fmt"
import "strconv"

type parameters struct {
	fields        int
	maxWordLength int
	minWordLength int
	lineCount     int
	noComma       bool
}

func (p *parameters) String() string {
	return fmt.Sprintf("parameters(lines=%v, fields=%v, maxWordLength=%v, minWordLength=%v)",
		p.lineCount, p.fields, p.maxWordLength, p.minWordLength)
}

func randomASCIIByte(parameters *parameters) byte {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := rand.Intn(95) + 32
	if parameters.noComma && number == 44 {
		number = 45
	}
	return byte(number)
}

func randomWord(parameters *parameters) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	length := rand.Intn(parameters.maxWordLength-parameters.minWordLength) + parameters.minWordLength

	characters := make([]byte, length)
	for index := 0; index < length; index += 1 {
		characters[index] = randomASCIIByte(parameters)
	}

	return string(characters)
}

func randomLine(parameters *parameters) string {
	randomWords := make([]string, parameters.fields)
	for index := 0; index < parameters.fields; index += 1 {
		randomWords[index] = randomWord(parameters)
	}
	return strings.Join(randomWords, ",")
}

func randomFile(parameters *parameters) string {
	randomLines := make([]string, parameters.lineCount)
	for index := 0; index < parameters.fields; index += 1 {
		randomLines[index] = randomLine(parameters)
	}

	return strings.Join(randomLines, "\n")
}

func parseArguments(arguments []string) *parameters {

	lineCount := 0
	fieldCount := 0
	minWordLength := 0
	maxWordLength := 10
	noComma := false

	for _, argument := range arguments {
		switch {
		case strings.HasPrefix(argument, "-l"):
			number, _ := strconv.ParseInt(argument[2:], 10, 32)
			lineCount = int(number)

		case strings.HasPrefix(argument, "-f"):
			number, _ := strconv.ParseInt(argument[2:], 10, 32)
			fieldCount = int(number)

		case strings.HasPrefix(argument, "-cmax"):
			number, _ := strconv.ParseInt(argument[5:], 10, 32)
			maxWordLength = int(number)

		case strings.HasPrefix(argument, "-cmin"):
			number, _ := strconv.ParseInt(argument[5:], 10, 32)
			minWordLength = int(number)

		case argument == "--no-comma-values":
			noComma = true
		}
	}

	return &parameters{
		lineCount:     lineCount,
		fields:        fieldCount,
		minWordLength: minWordLength,
		maxWordLength: maxWordLength,
		noComma:       noComma,
	}
}

func gen(arguments []string, output io.Writer) {
	parameters := parseArguments(arguments)
	for lineCount := 1; lineCount <= parameters.lineCount; lineCount += 1 {
		io.WriteString(output, fmt.Sprintln(randomLine(parameters)))
	}
}

func main() {
	gen(os.Args[1:], os.Stdout)
}
