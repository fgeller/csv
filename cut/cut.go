package main

import "os"
import "bytes"
import "fmt"
import "bufio"
import "io"
import "strings"
import "strconv"
import "runtime/pprof"
import "log"

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
	mode := csvMode
	inputDelimiter := ""
	outputDelimiter := ""
	fileNames := []string{}
	delimitedOnly := false
	complement := false
	headerNames := ""
	lineEnd := ""
	cpuProfile := false
	printUsage := false
	printVersion := false

	for index := 0; index < len(rawArguments); index += 1 {
		argument := rawArguments[index]
		switch {

		case argument == "-b" || argument == "--bytes":
			mode = byteMode
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-b"):
			mode = byteMode
			ranges = argument[len("-b"):]
		case strings.HasPrefix(argument, "--bytes="):
			mode = byteMode
			ranges = argument[len("--bytes="):]

		case argument == "-c" || argument == "--characters":
			mode = characterMode
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-c"):
			mode = characterMode
			ranges = argument[len("-c"):]
		case strings.HasPrefix(argument, "--characters="):
			mode = characterMode
			ranges = argument[len("--characters="):]

		case argument == "-d" || argument == "--delimiter":
			inputDelimiter = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-d"):
			inputDelimiter = argument[len("-d"):]
		case strings.HasPrefix(argument, "--delimiter="):
			inputDelimiter = argument[len("--delimiter="):]

		case argument == "-e":
			mode = csvMode
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-e"):
			mode = csvMode
			ranges = argument[2:]

			// case argument == "-E":
			//	mode = csvMode
			//	headerNames = rawArguments[index+1]
			//	index += 1
			// case strings.HasPrefix(argument, "-E"):
			//	mode = csvMode
			//	headerNames = argument[2:]

		case argument == "-f" || argument == "--fields":
			mode = fieldMode
			ranges = rawArguments[index+1]
			index += 1
		case strings.HasPrefix(argument, "-f"):
			mode = fieldMode
			ranges = argument[len("-f"):]
		case strings.HasPrefix(argument, "--fields="):
			mode = fieldMode
			ranges = argument[len("--fields="):]

		case argument == "-n":
			// ignore

		case argument == "--complement":
			complement = true

		case argument == "-s" || argument == "--only-delimited":
			delimitedOnly = true

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
		return nil, fmt.Sprintf("%s", err)
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
	if len(parameters.ranges) == 0 {
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

	return false
}

func cutCSV(input io.Reader, output io.Writer, parameters *parameters) {
	bufferedInput := bufio.NewReaderSize(input, 4096)
	bufferedOutput := bufio.NewWriterSize(output, 4096)
	defer bufferedOutput.Flush()

	buffer := make([]byte, 4096*1000)
	word := make([]byte, 0, 30)
	selected := make([]bool, 0, 20)

	inputDelimiter := []byte(parameters.inputDelimiter)
	inputDelimiterEndByte := inputDelimiter[len(inputDelimiter)-1]

	outputDelimiter := []byte(parameters.outputDelimiter)

	lineEnd := []byte(parameters.lineEnd)
	lineEndByte := lineEnd[len(lineEnd)-1]

	inEscaped := false
	inHeader := true
	haveSeenDelimiter := false
	firstWordWritten := false
	wordCount := 1
	mode := parameters.mode

	writeOut := func(eol bool) bool {
		if inHeader {
			selected = append(selected, isSelected(parameters, wordCount))
		}

		force := eol && !haveSeenDelimiter && !parameters.delimitedOnly

		if force || selected[wordCount-1] && (haveSeenDelimiter || !parameters.delimitedOnly) {
			if firstWordWritten {
				bufferedOutput.Write(outputDelimiter)
			}

			bufferedOutput.Write(word)
			firstWordWritten = true
			word = word[:0]
			return true
		}

		word = word[:0]
		return false
	}

	for {
		count, err := bufferedInput.Read(buffer)

		for bufferIndex := 0; bufferIndex < count; bufferIndex += 1 {
			char := buffer[bufferIndex]

			switch {

			case mode == csvMode && !inEscaped && char == DQUOTE:
				inEscaped = true
				word = append(word, char)

			case mode == csvMode && inEscaped && char == DQUOTE:
				inEscaped = false
				word = append(word, char)

			case mode == csvMode && !inEscaped && char == inputDelimiterEndByte || mode == fieldMode && char == inputDelimiterEndByte:
				word = append(word, char)
				if bytes.Equal(word[len(word)-len(inputDelimiter):], inputDelimiter) {
					word = word[:len(word)-len(inputDelimiter)]
					haveSeenDelimiter = true
					writeOut(false)
					wordCount += 1
				}

			case !inEscaped && char == lineEndByte:
				word = append(word, char)
				if bytes.Equal(word[len(word)-len(lineEnd):], lineEnd) {
					word = word[:len(word)-len(lineEnd)]
					writeOut(true)
					if firstWordWritten {
						bufferedOutput.Write(lineEnd)
					}
					inHeader = false
					haveSeenDelimiter = false
					firstWordWritten = false
					wordCount = 1
				}

			case true:
				word = append(word, char)
			}

		}

		if err != nil {
			if len(word) > 0 {
				writeOut(true)
				wordCount = 0
				bufferedOutput.Write(lineEnd)
			}
			break
		}
	}
}

func cutBytes(input io.Reader, output io.Writer, parameters *parameters) {
	bufferedInput := bufio.NewReaderSize(input, 4096)
	bufferedOutput := bufio.NewWriterSize(output, 4096)
	defer bufferedOutput.Flush()

	buffer := make([]byte, 4096*1000)

	firstWrittenByte := true
	shouldInsertSeparator := len(parameters.outputDelimiter) > 0
	separator := []byte(parameters.outputDelimiter)

	inHeader := true
	selected := make([]bool, 0, 20)
	byteCount := 1

	for {
		count, err := bufferedInput.Read(buffer)

		for bufferIndex := 0; bufferIndex < count; bufferIndex += 1 {
			char := buffer[bufferIndex]

			if inHeader {
				selected = append(selected, isSelected(parameters, byteCount))
			}

			if selected[byteCount-1] {
				if shouldInsertSeparator && !firstWrittenByte {
					bufferedOutput.Write(separator)
				}
				firstWrittenByte = false
				bufferedOutput.WriteByte(char)
			}

			if char == LF {
				inHeader = false
				firstWrittenByte = true
				byteCount = 1
			} else {
				if shouldInsertSeparator && selected[byteCount-1] {
					bufferedOutput.Write(separator)
				}
				byteCount += 1
			}
		}

		if err != nil {
			bufferedOutput.WriteByte(LF)
			break
		}
	}
}

func cutCharacters(input io.Reader, output io.Writer, parameters *parameters) {
	bufferedInput := bufio.NewReaderSize(input, 4096)
	bufferedOutput := bufio.NewWriterSize(output, 4096)
	defer bufferedOutput.Flush()

	firstWrittenRune := true
	shouldInsertSeparator := len(parameters.outputDelimiter) > 0
	separator := []byte(parameters.outputDelimiter)

	inHeader := true
	selected := make([]bool, 0, 20)
	runeCount := 1

	for {
		rune, size, err := bufferedInput.ReadRune()

		if size > 0 && err == nil {
			if inHeader {
				selected = append(selected, isSelected(parameters, runeCount))
			}

			if selected[runeCount-1] {
				if shouldInsertSeparator && !firstWrittenRune {
					bufferedOutput.Write(separator)
				}
				firstWrittenRune = false
				bufferedOutput.WriteRune(rune)
			}

			if rune == '\n' {
				inHeader = false
				firstWrittenRune = true
				runeCount = 1
			} else {
				runeCount += 1
			}

		} else {
			bufferedOutput.WriteByte(LF)
			break
		}
	}
}

func cutFile(input io.Reader, output io.Writer, parameters *parameters) {
	switch {
	case parameters.mode == characterMode:
		cutCharacters(input, output, parameters)
	case parameters.mode == byteMode:
		cutBytes(input, output, parameters)
	case parameters.mode == fieldMode || parameters.mode == csvMode:
		cutCSV(input, output, parameters)
	}
}

func printUsage(output io.Writer) {
	usage := `Usage: cut OPTION... [FILE]...
Print selected parts of lines from each file to standard output.

Mandatory arguments to long options are mandatory for short options too.
  -b, --bytes=LIST        select only these bytes
  -c, --characters=LIST   select only these characters
  -d, --delimiter=DELIM   use DELIM instead of TAB for field delimiter
  -e LIST                 select only comma separated columns
  -f, --fields=LIST       select only these fields;  also print any line
                            that contains no delimiter character, unless
                            the -s option is specified
  -n                      (ignored)
      --complement        complement the set of selected bytes, characters
                            or fields
  -s, --only-delimited    do not print lines not containing delimiters
      --output-delimiter=STRING  use STRING as the output delimiter
                            the default is to use the input delimiter
      --help     display this help and exit
      --version  output version information and exit

Use one, and only one of -b, -c or -f.  Each LIST is made up of one
range, or many ranges separated by commas.  Selected input is written
in the same order that it is read, and is written exactly once.
Each range is one of:

  N     N'th byte, character or field, counted from 1
  N-    from N'th byte, character or field, to end of line
  N-M   from N'th to M'th (included) byte, character or field
  -M    from first to M'th (included) byte, character or field

With no FILE, or when FILE is -, read standard input.

The project is available online at https://github.com/fgeller/csv-cut

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
		fmt.Printf("CPU profiling output will be written to cut.cprof\n")
		f, err := os.Create("cut.cprof")
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
