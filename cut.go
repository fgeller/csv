package main

import "os"
import "fmt"
import "bufio"
import "io"
import "strings"
import "strconv"

const (
	fields_message    string = "select only these fields"
	delimiter_message string = "custom delimiter"
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
	ranges        []Range
	delimiter     string
	delimitedOnly bool
	input         []*os.File
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
	// default values
	ranges := ""
	delimiter := ","
	fileNames := []string{}
	delimitedOnly := false

	for index := 0; index < len(rawArguments); index += 1 {
		argument := rawArguments[index]
		switch {
		case argument == "-f":
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-f"):
			ranges = argument[2:]

		case argument == "-d":
			delimiter = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-d"):
			delimiter = argument[2:]

		case argument == "-s" || argument == "--only-delimited":
			delimitedOnly = true

		case true:
			fileNames = append(fileNames, argument)
		}
	}

	input, err := openInput(fileNames)
	if err != nil {
		return nil, err
	}

	return &parameters{
		ranges:        parseRanges(ranges),
		delimiter:     delimiter,
		input:         input,
		delimitedOnly: delimitedOnly,
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

// TODO: can we do this lazily?
func selectedFields(ranges []Range, total int) []int {
	selected := make([]int, 0)
outer:
	for field := 1; field <= total; field += 1 {
		for _, aRange := range ranges {
			if aRange.Contains(field) {
				selected = append(selected, field)
				continue outer
			}
		}
	}

	return selected
}

func collectFields(fields []string, selected []int) []string {
	if 0 == len(selected) {
		return fields
	}

	collectedFields := make([]string, len(selected))
	for index, selectedField := range selected {
		collectedFields[index] = fields[selectedField-1]
	}

	return collectedFields
}

func cutLine(line string, parameters *parameters) string {
	if !strings.Contains(line, parameters.delimiter) {
		return line
	}

	fields := strings.Split(line, parameters.delimiter)
	selected := selectedFields(parameters.ranges, len(fields))
	collectedFields := collectFields(fields, selected)

	return strings.Join(collectedFields, parameters.delimiter)
}

func skipLine(line string, parameters *parameters) bool {
	return len(line) > 0 &&
		parameters.delimitedOnly &&
		!strings.Contains(line, parameters.delimiter)
}

func ensureNewLine(line string) string {
	return fmt.Sprintln(strings.TrimSuffix(line, "\n"))
}

func cutFile(input io.Reader, output io.Writer, parameters *parameters) {
	reader := bufio.NewReader(input)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Println("Encountered error while reading:", err)
		}

		if !skipLine(line, parameters) {
			newLine := ensureNewLine(cutLine(line, parameters))
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

// -b, --bytes=LIST
//        select only these bytes
// -c, --characters=LIST
//        select only these characters
// -d, --delimiter=DELIM
//        use DELIM instead of TAB for field delimiter
// -f, --fields=LIST
//        select only these fields;  also print any line that contains no delimiter character, unless the -s option is specified
// -n     (ignored)
// --complement
//        complement the set of selected bytes, characters or fields
// -s, --only-delimited
//        do not print lines not containing delimiters
// --output-delimiter=STRING
//        use STRING as the output delimiter the default is to use the input delimiter
// --help display this help and exit
// --version
//        output version information and exit

func main() {
	cut(os.Args[1:], os.Stdout)
}
