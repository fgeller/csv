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

type parameters struct {
    ranges    []*Range
    delimiter string
    input     []*os.File
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
        case true:
            fileNames = append(fileNames, argument)
        }
    }

    input, err := openInput(fileNames)
    if err != nil {
        return nil, err
    }

    return &parameters{
        ranges:    parseRanges(ranges),
        delimiter: delimiter,
        input:     input,
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

func parseRanges(rawRanges string) []*Range {
    if 0 == len(rawRanges) {
        return []*Range{}
    }

    splits := strings.Split(rawRanges, ",")
    ranges := make([]*Range, 0)

    for _, split := range splits {
        number, _ := strconv.ParseInt(split, 10, 32)
        ranges = append(ranges, &Range{start: int(number)})
    }

    return ranges
}

func parseFields(selectedFields string) []int64 {
    if 0 == len(selectedFields) {
        return []int64{}
    }

    splits := strings.Split(selectedFields, ",")
    numbers := make([]int64, len(splits))
    for idx, stringNumber := range splits {
        numbers[idx], _ = strconv.ParseInt(stringNumber, 10, 64)
    }

    return numbers
}

func collectFields(fields []string, ranges []*Range) []string {
    if 0 == len(ranges) {
        return fields
    }

    collectedFields := make([]string, 0)
    for _, aRange := range ranges {
        index := aRange.start
        if index == -1 { // hacky hack
            return fields
        }
        field := fields[index-1]
        collectedFields = append(collectedFields, field)
    }
    return collectedFields
}

// TODO take param with modes
func cutFile(input io.Reader, output io.Writer, delimiter string, ranges []*Range) {
    reader := bufio.NewReader(input)

    for {
        line, err := reader.ReadString('\n')
        fields := strings.Split(line, delimiter)

        collectedFields := collectFields(fields, ranges)

        newLine := fmt.Sprintln(strings.TrimSuffix(strings.Join(collectedFields, delimiter), "\n"))
        _, writeErr := io.WriteString(output, newLine)

        if err == io.EOF {
            break
        }

        if err != nil {
            fmt.Println("Encountered error while reading:", err)
            break
        }

        if writeErr != nil {
            fmt.Println("Encountered error while writing:", writeErr)
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
        cutFile(file, output, parameters.delimiter, parameters.ranges)
    }
}

func main() {
    cut(os.Args[1:], os.Stdout)
}
