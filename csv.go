package main

import "os"
import "fmt"
import "bufio"
import "io"
import "strings"
import "strconv"
import "runtime/pprof"
import "log"

const (
	DQUOTE byte = 0x22
	COMMA  byte = 0x2c
	CR     byte = 0x0d
	LF     byte = 0x0a
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
	ranges          []Range
	inputDelimiter  string
	outputDelimiter string
	complement      bool
	input           []*os.File
	names           []string
	lineEnd         string
	cpuProfile      bool
	printUsage      bool
	printVersion    bool
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

func parseArguments(rawArguments []string) (*parameters, string) {
	ranges := ""
	inputDelimiter := ""
	outputDelimiter := ""
	fileNames := []string{}
	complement := false
	names := ""
	lineEnd := ""
	cpuProfile := false
	printUsage := false
	printVersion := false

	for index := 0; index < len(rawArguments); index += 1 {
		argument := rawArguments[index]
		switch {

		case argument == "-d" || argument == "--delimiter":
			inputDelimiter = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-d"):
			inputDelimiter = argument[len("-d"):]
		case strings.HasPrefix(argument, "--delimiter="):
			inputDelimiter = argument[len("--delimiter="):]

		case argument == "-c" || argument == "--columns":
			ranges = rawArguments[index+1]
			index += 1
			lineEnd = string(LF)
		case strings.HasPrefix(argument, "-c"):
			ranges = argument[len("-c"):]
			lineEnd = string(LF)
		case strings.HasPrefix(argument, "--columns="):
			ranges = argument[len("--columns="):]
			lineEnd = string(LF)

		case argument == "-C" || argument == "--Columns":
			ranges = rawArguments[index+1]
			index += 1
			lineEnd = string([]byte{CR, LF})
		case strings.HasPrefix(argument, "-C"):
			ranges = argument[len("-C"):]
			lineEnd = string([]byte{CR, LF})
		case strings.HasPrefix(argument, "--Columns="):
			ranges = argument[len("--Columns="):]
			lineEnd = string([]byte{CR, LF})

		case argument == "-n" || argument == "--names":
			names = rawArguments[index+1]
			index += 1
			lineEnd = string(LF)
		case strings.HasPrefix(argument, "-n"):
			names = argument[len("-n"):]
			lineEnd = string(LF)
		case strings.HasPrefix(argument, "--names="):
			names = argument[len("--names="):]
			lineEnd = string(LF)

		case argument == "-N" || argument == "--Names":
			names = rawArguments[index+1]
			index += 1
			lineEnd = string([]byte{CR, LF})
		case strings.HasPrefix(argument, "-N"):
			names = argument[len("-N"):]
			lineEnd = string([]byte{CR, LF})
		case strings.HasPrefix(argument, "--Names="):
			names = argument[len("--Names="):]
			lineEnd = string([]byte{CR, LF})

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

		case argument == "--cpuprofile":
			cpuProfile = true

		case argument == "--help":
			printUsage = true

		case argument == "--version":
			printVersion = true

		case argument == "-":
			fileNames = nil

		case strings.HasPrefix(argument, "-"):
			return nil, fmt.Sprintf("Invalid argument %s", argument)

		case true:
			fileNames = append(fileNames, argument)
		}
	}

	if inputDelimiter == "" {
		inputDelimiter = ","
	}

	if len(outputDelimiter) == 0 {
		outputDelimiter = inputDelimiter
	}

	if len(lineEnd) == 0 {
		lineEnd = string(LF)
	}

	input, err := openInput(fileNames)
	if err != nil {
		return nil, fmt.Sprintf("%s", err)
	}

	return &parameters{
		ranges:          parseRanges(ranges),
		names:           parseNames(names),
		inputDelimiter:  inputDelimiter,
		outputDelimiter: outputDelimiter,
		input:           input,
		complement:      complement,
		lineEnd:         lineEnd,
		cpuProfile:      cpuProfile,
		printUsage:      printUsage,
		printVersion:    printVersion,
	}, ""
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

func parseNames(rawNames string) []string {
	if len(rawNames) == 0 {
		return []string{}
	}
	return strings.Split(rawNames, ",")
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

func isSelected(parameters *parameters, field int, word string) bool {
	if len(parameters.ranges) == 0 && len(parameters.names) == 0 {
		return true
	}

	for _, aRange := range parameters.ranges {
		contained := aRange.Contains(field)
		switch {
		case !parameters.complement && contained:
			return true
		case parameters.complement && !contained:
			return true
		}
	}

	for _, name := range parameters.names {
		wrappedName := fmt.Sprintf("%v%v%v", string(DQUOTE), name, string(DQUOTE))

		switch {
		case !parameters.complement && (name == word || wrappedName == word):
			return true
		case parameters.complement && (name != word && wrappedName != word):
			return true
		}
	}

	return false
}

func cutFile(input io.Reader, output io.Writer, parameters *parameters) {
	bufferedInput := bufio.NewReaderSize(input, 4096)
	bufferedOutput := bufio.NewWriterSize(output, 4096)
	defer bufferedOutput.Flush()

	inputDelimiter := []byte(parameters.inputDelimiter)
	outputDelimiter := []byte(parameters.outputDelimiter)

	lineEnd := []byte(parameters.lineEnd)

	inEscaped := false

	inputDelimiterIndex := 0
	inInputDelimiter := false
	lineEndIndex := 0
	inLineEnd := true

charLoop:
	for {
		char, err := bufferedInput.ReadByte()
		if err != nil {
			break
		}

		// TODO: save first line to set selected for selection by header

		if inInputDelimiter && inputDelimiterIndex == len(inputDelimiter) {
			bufferedOutput.Write(outputDelimiter)
			inInputDelimiter = false
			inputDelimiterIndex = 0
		}
		if inLineEnd && lineEndIndex == len(lineEnd) {
			bufferedOutput.Write(lineEnd)
			inLineEnd = false
			lineEndIndex = 0
		}

		switch {
		case !inEscaped && char == DQUOTE:
			inEscaped = true

		case inEscaped && char == DQUOTE:
			inEscaped = false
		}

		if !inEscaped && char == inputDelimiter[inputDelimiterIndex] {
			inInputDelimiter = true
			inputDelimiterIndex += 1
			continue charLoop
		}

		if !inEscaped && char == lineEnd[lineEndIndex] {
			inLineEnd = true
			lineEndIndex += 1
			continue charLoop
		}

		inputDelimiterIndex = 0
		lineEndIndex = 0
		bufferedOutput.WriteByte(char);

	}
}

func printUsage(output io.Writer) {
	usage := `Usage: csv OPTION... [FILE]...
Print selected comma separated values of lines from each file to standard output.

Mandatory arguments to long options are mandatory for short options too.
  -c, --columns=LIST             select columns by position, line ending LF
  -C, --Columns=LIST             select columns by position, line ending CRLF
  -n, --names=NAMES              select columns by name, line ending LF
  -N, --Names=NAMES              select columns by name, line ending CRLF
  -d, --delimiter=DELIM          use DELIM instead of TAB for field delimiter
      --complement               complement the set of columns
      --output-delimiter=STRING  use STRING as the output delimiter
                                 the default is to use the input delimiter
      --help                     display this help and exit
      --version                  output version information and exit

Each NAMES is made up of a list of names separated by commas. The first line of
each file is interpreted as the header line in which names are defined.

Each LIST is made up of one range, or many ranges separated by commas.  Selected
input is written in the same order that it is read, and is written exactly once.
Each range is one of:

  N     N'th byte, character or field, counted from 1
  N-    from N'th byte, character or field, to end of line
  N-M   from N'th to M'th (included) byte, character or field
  -M    from first to M'th (included) byte, character or field

With no FILE, or when FILE is -, read standard input.

The project is available online at https://github.com/fgeller/csv

Credits:
As the interface is based on cut from GNU coreutils, much of this usage
information is taken from taken from GNU coreutils version.

GNU coreutils is available at: <http://www.gnu.org/software/coreutils/>
`
	output.Write([]byte(usage))
}

func printVersion(output io.Writer) {
	usage := `cut 0.314
`
	output.Write([]byte(usage))
}

func printInvalidUsage(output io.Writer, message string) {
	usage := fmt.Sprintf(`%v: %v
Try '%s --help' for more information.
`, os.Args[0], message, os.Args[0])
	output.Write([]byte(usage))
}

func cut(arguments []string, output io.Writer) {
	parameters, err := parseArguments(arguments)
	if err != "" {
		printInvalidUsage(os.Stderr, err)
		return
	}
	if parameters.printUsage {
		printUsage(output)
		return
	}
	if parameters.printVersion {
		printVersion(output)
		return
	}

	if parameters.cpuProfile {
		fmt.Printf("CPU profiling output will be written to csv.cprof\n")
		f, err := os.Create("csv.cprof")
		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	for _, file := range parameters.input {
		cutFile(file, output, parameters)
	}
}

func main() {
	cut(os.Args[1:], os.Stdout)
}
