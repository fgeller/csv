package main

import "os"
import "fmt"
import "bufio"
import "io"
import "strings"
import "strconv"
import "runtime/pprof"

const (
	fields_message    string = "select only these fields"
	delimiter_message string = "custom delimiter"
)

const (
	DQUOTE byte = 0x22
	COMMA  byte = 0x2c
	CR     byte = 0x0d
	LF     byte = 0x0a
)

const (
	fieldMode = iota
	byteMode
	characterMode
	csvMode
)

// mhh... how to resolve naming conflict?
type Range struct {
	start int
	end   int
}

func (r Range) Contains(number int) bool {
	switch {
	case r.start == 0 && number <= r.end:
		return true
	case r.end == 0 && r.start <= number:
		return true
	case r.start <= number && number <= r.end:
		return true
	}
	return false
}

func (r Range) String() string {
	return fmt.Sprintf("Range(%v, %v)", r.start, r.end)
}

func NewRange(start int, end int) Range {
	return Range{start: start, end: end}
}

type parameters struct {
	mode            int
	ranges          []Range
	inputDelimiter  string
	outputDelimiter string
	delimitedOnly   bool
	complement      bool
	input           []*os.File
	headerNames     string
	lineEnd         string
}

func openInput(fileNames []string) ([]*os.File, error) {
	if 0 == len(fileNames) || fileNames[0] == "-" {
		return []*os.File{os.Stdin}, nil
	}

	opened, err := openFiles(fileNames)
	if err != nil {
		return nil, err
	}

	return opened, nil
}

// TODO: -d" "
func parseArguments(rawArguments []string) (*parameters, error) {
	ranges := ""
	mode := fieldMode
	inputDelimiter := ""
	outputDelimiter := ""
	fileNames := []string{}
	delimitedOnly := false
	complement := false
	headerNames := ""
	lineEnd := ""

	for index := 0; index < len(rawArguments); index += 1 {
		argument := rawArguments[index]
		switch {

		case argument == "-f":
			mode = fieldMode
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-f"):
			mode = fieldMode
			ranges = argument[2:]

		case argument == "-e":
			mode = csvMode
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-e"):
			mode = csvMode
			ranges = argument[2:]

		case argument == "-E":
			mode = csvMode
			headerNames = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-E"):
			mode = csvMode
			headerNames = argument[2:]

		case argument == "-c":
			mode = characterMode
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-c"):
			mode = characterMode
			ranges = argument[2:]

		case argument == "-b":
			mode = byteMode
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-b"):
			mode = byteMode
			ranges = argument[2:]

		case argument == "-d":
			inputDelimiter = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-d"):
			inputDelimiter = argument[2:]

		case argument == "-s" || argument == "--only-delimited":
			delimitedOnly = true

		case argument == "--complement":
			complement = true

		case argument == "--output-delimiter":
			outputDelimiter = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "--output-delimiter="):
			outputDelimiter = argument[19:]

		case strings.HasPrefix(argument, "--line-end="):
			switch {
			case argument[11:] == "LF":
				lineEnd = string(LF)
			case argument[11:] == "CRLF":
				lineEnd = string([]byte{CR, LF})
			}

		case argument == "-n":
			// ignore

		case true:
			fileNames = append(fileNames, argument)
		}
	}

	switch {
	case inputDelimiter == "" && mode == csvMode:
		inputDelimiter = ","
	case inputDelimiter == "":
		inputDelimiter = "\t"
	}

	if len(outputDelimiter) == 0 && (mode == fieldMode || mode == csvMode) {
		outputDelimiter = inputDelimiter
	}

	switch {
	case len(lineEnd) == 0 && mode == csvMode:
		lineEnd = string([]byte{CR, LF})
	case len(lineEnd) == 0:
		lineEnd = "\n"
	}

	input, err := openInput(fileNames)
	if err != nil {
		return nil, err
	}

	return &parameters{
		mode:            mode,
		ranges:          parseRanges(ranges),
		inputDelimiter:  inputDelimiter,
		outputDelimiter: outputDelimiter,
		input:           input,
		delimitedOnly:   delimitedOnly,
		complement:      complement,
		headerNames:     headerNames,
		lineEnd:         lineEnd,
	}, nil
}

func openFiles(fileNames []string) ([]*os.File, error) {
	files := make([]*os.File, len(fileNames))

	for index, fileName := range fileNames {
		file, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}

		files[index] = file
	}

	return files, nil
}

func parseInt(raw string) int {
	number, _ := strconv.ParseInt(raw, 10, 32)
	return int(number)
}

func parseRange(raw string) Range {
	splitPosition := strings.Index(raw, "-")

	if splitPosition == -1 {
		number := parseInt(raw)
		return NewRange(number, number)
	}

	lower := raw[:splitPosition]
	upper := raw[splitPosition+1:]

	return NewRange(parseInt(lower), parseInt(upper))
}

func parseRanges(rawRanges string) []Range {
	if 0 == len(rawRanges) {
		return []Range{}
	}

	ranges := make([]Range, 0)
	for _, raw := range strings.Split(rawRanges, ",") {
		ranges = append(ranges, parseRange(raw))
	}

	return ranges
}

func isSelected(parameters *parameters, field int) bool {
	for _, aRange := range parameters.ranges {
		contained := aRange.Contains(field)
		switch {
		case !parameters.complement && contained:
			return true
		case parameters.complement && !contained:
			return true
		}
	}

	return false
}

// TODO: can we do this lazily?
func selectedFields(parameters *parameters, total int) []int {
	selected := make([]int, 0)
outer:
	for field := 1; field <= total; field += 1 {
		for _, aRange := range parameters.ranges {
			contained := aRange.Contains(field)
			switch {
			case !parameters.complement && contained:
				selected = append(selected, field)
				continue outer
			case parameters.complement && !contained:
				selected = append(selected, field)
				continue outer
			}
		}
	}

	return selected
}

func collectCharacters(line string, parameters *parameters) []string {
	runes := []rune(line)
	selected := selectedFields(parameters, len(runes))
	if 0 == len(selected) {
		return []string{line}
	}

	collectedCharacters := make([]string, len(selected))
	for index, selectedField := range selected {
		collectedCharacters[index] = string(runes[selectedField-1 : selectedField])
	}

	return collectedCharacters
}

func collectBytes(line string, parameters *parameters) []string {
	bytes := []byte(line)
	selected := selectedFields(parameters, len(bytes))
	if 0 == len(selected) {
		return []string{line}
	}

	collectedCharacters := make([]string, len(selected))
	for index, selectedField := range selected {
		collectedCharacters[index] = string(bytes[selectedField-1 : selectedField])
	}

	return collectedCharacters
}

func collectFields(line string, parameters *parameters) []string {
	if !strings.Contains(line, parameters.inputDelimiter) {
		return []string{line}
	}

	fields := strings.Split(line, parameters.inputDelimiter)
	selected := selectedFields(parameters, len(fields))

	if 0 == len(selected) {
		return []string{line}
	}

	collectedFields := make([]string, len(selected))
	for index, selectedField := range selected {
		collectedFields[index] = fields[selectedField-1]
	}

	return collectedFields
}

func cutCSVFields(input *bufio.Reader, output *bufio.Writer, parameters *parameters) {
	outputDelimiter := []byte(parameters.outputDelimiter)
	lineEnd := []byte(parameters.lineEnd)
	inEscaped := false
	eolMatchIndex := 0
	eolMatch := make([]bool, len(lineEnd))

	resetEolMatch := func() {
		// TODO
		// if eolMatchIndex > 0 && eolMatchIndex < len(lineEnd) {
		//	// fmt.Printf("------Should flush out %s bytes.\n", eolMatchIndex)
		// }

		eolMatchIndex = 0
		for index := range eolMatch {
			eolMatch[index] = false
		}
	}

	wordCount := 1
	writeOut := func(char byte) {
		if isSelected(parameters, wordCount) { // cache the selection?
			output.WriteByte(char)
		}
	}

	finishWord := func() {
		if isSelected(parameters, wordCount) { // cache the selection?
			output.Write(outputDelimiter)
		}
		wordCount += 1
	}

	finishLine := func() {
		output.Write(lineEnd)
		wordCount = 1
	}

	for {
		char, err := input.ReadByte()

		if err == io.EOF || eolMatchIndex == len(lineEnd) {
			finishLine()
			resetEolMatch()
		}
		if err == io.EOF {
			break
		}

		if !inEscaped && char == lineEnd[eolMatchIndex] {
			eolMatch[eolMatchIndex] = true
			eolMatchIndex += 1
		} else {
			resetEolMatch()
		}

		switch {

		case !inEscaped && char == DQUOTE:
			inEscaped = true
			writeOut(char)

		case inEscaped && char == DQUOTE:
			inEscaped = false
			writeOut(char)

		case !inEscaped && char == COMMA:
			finishWord()

		case !inEscaped && eolMatchIndex == len(lineEnd): // at EOL
			continue

		case err == nil:
			if eolMatchIndex > 0 {
			} else {
				writeOut(char)
			}

		}

	}

}

func pickSelected(fields [][]byte, selected []int, parameters *parameters) [][]byte {
	if 0 == len(selected) || 0 == len(fields) {
		return fields
	}

	collectedFields := make([][]byte, len(selected))
	for index, selectedField := range selected {
		collectedFields[index] = fields[selectedField-1]
	}

	return collectedFields
}

func cutLine(line string, parameters *parameters) string {
	collectedFields := []string{}
	switch {
	case parameters.mode == fieldMode:
		collectedFields = collectFields(line, parameters)
	case parameters.mode == byteMode:
		collectedFields = collectBytes(line, parameters)
	case parameters.mode == characterMode:
		collectedFields = collectCharacters(line, parameters)
	}

	return strings.Join(collectedFields, parameters.outputDelimiter)
}

func skipLine(line string, parameters *parameters) bool {
	return len(line) > 0 &&
		parameters.delimitedOnly &&
		!strings.Contains(line, parameters.inputDelimiter)
}

func ensureNewLine(line string, lineEnd string) string {
	return fmt.Sprintf("%v%v", strings.TrimSuffix(line, lineEnd), lineEnd)
}

func selectFieldsByName(parameters *parameters, headers []string) []int {
	rawNames := strings.Split(parameters.headerNames, ",")
	selectedNames := make([]string, len(rawNames))
	for index, rawName := range rawNames {
		selectedNames[index] = strings.Trim(rawName, "\"")
	}

	selected := make([]int, 0)
	for index, header := range headers {
		for _, selectedName := range selectedNames {
			if selectedName == header {
				selected = append(selected, index+1) // :|
			}
		}
	}

	return selected
}

func cutCSVFile(input io.Reader, output io.Writer, parameters *parameters) {
	bufferedInput := bufio.NewReaderSize(input, 1024*1024)
	bufferedOutput := bufio.NewWriterSize(output, 1024*1024)
	defer bufferedOutput.Flush()

	cutCSVFields(bufferedInput, bufferedOutput, parameters)
}

func cutFile(input io.Reader, output io.Writer, parameters *parameters) {

	// :|
	if parameters.mode == csvMode {
		cutCSVFile(input, output, parameters)
		return
	}

	reader := bufio.NewReader(input)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Println("Encountered error while reading:", err)
		}

		if !skipLine(line, parameters) {
			newLine := ensureNewLine(cutLine(line, parameters), parameters.lineEnd)
			_, writeErr := io.WriteString(output, newLine)
			if writeErr != nil {
				fmt.Println("Encountered error while writing:", writeErr)
				break
			}
		}

		if err != nil {
			break
		}
	}
}

func cut(arguments []string, output io.Writer) {
	parameters, err := parseArguments(arguments)
	if err != nil {
		fmt.Println("Invalid arguments:", err)
		return
	}

	for _, file := range parameters.input {
		cutFile(file, output, parameters)
	}
}

func main() {
	f, _ := os.Create("cut.cprof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	cut(os.Args[1:], os.Stdout)
}
