package main

import "io"
import "strconv"
import "strings"
import "os"
import "bufio"

type parameters struct {
	copy         bool
	inputFiles   []*os.File
	inputBuffer  int
	outputBuffer int
	chunkSize    int
}

func parseArguments(arguments []string) *parameters {
	inputFiles := []*os.File{}
	inputBuffer := 1
	outputBuffer := 1
	chunkSize := 1
	copy := false

	for _, argument := range arguments {
		switch {
		case argument == "--copy":
			copy = true

		case strings.HasPrefix(argument, "--chunks="):
			number, _ := strconv.ParseInt(argument[len("--chunks="):], 10, 64)
			chunkSize = int(number)

		case strings.HasPrefix(argument, "--inputBuffer="):
			number, _ := strconv.ParseInt(argument[len("--inputBuffer="):], 10, 64)
			inputBuffer = int(number)

		case strings.HasPrefix(argument, "--outputBuffer="):
			number, _ := strconv.ParseInt(argument[len("--outputBuffer="):], 10, 64)
			outputBuffer = int(number)

		case true:
			file, _ := os.Open(argument)
			inputFiles = append(inputFiles, file)
		}
	}

	return &parameters{
		inputFiles:   inputFiles,
		copy:         copy,
		inputBuffer:  inputBuffer,
		outputBuffer: outputBuffer,
		chunkSize:    chunkSize,
	}

}

func cat(arguments []string, output io.Writer) {
	parameters := parseArguments(arguments)

inputs:
	for _, input := range parameters.inputFiles {

		if parameters.copy {
			io.Copy(output, input)
			continue inputs
		}

		bufferedInput := bufio.NewReaderSize(input, parameters.inputBuffer)
		bufferedOutput := bufio.NewWriterSize(output, parameters.outputBuffer)
		defer bufferedOutput.Flush()

		chunk := make([]byte, parameters.chunkSize)
		for {
			count, err := bufferedInput.Read(chunk)
			bufferedOutput.Write(chunk[:count])

			if err != nil {
				continue inputs
			}
		}
	}
}

func main() {
	cat(os.Args[1:], os.Stdout)
}
