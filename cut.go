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

type parameters struct {
    fields    []int64
    delimiter string
    files     []*os.File
}

func parseArguments(rawArguments []string) (*parameters, error) {
    // default values
    fields := ""
    delimiter := ","
    fileNames := []string{}

    for index := 0; index < len(rawArguments); index += 1 {
        argument := rawArguments[index]
        switch {
        case argument == "-f":
            fields = rawArguments[index+1]
            index += 1
        case strings.HasPrefix(argument, "-f"):
            fields = argument[2:]
        case argument == "-d":
            delimiter = rawArguments[index+1]
            index += 1
        case strings.HasPrefix(argument, "-d"):
            delimiter = argument[2:]
        case true:
            fileNames = append(fileNames, argument)
        }
    }

    files, err := openFiles(fileNames)
    if err != nil {
        return nil, err
    }

    return &parameters{
        fields:    parseFields(fields),
        delimiter: delimiter,
        files:     files,
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

func collectFields(fields []string, selectedFields []int64) []string {
    if 0 == len(selectedFields) {
        return fields
    }

    collectedFields := make([]string, len(selectedFields))
    for index, fieldIndex := range selectedFields {
        collectedFields[index] = fields[fieldIndex-1]
    }
    return collectedFields
}

func cut(input io.Reader, output io.Writer, delimiter string, selectedFields []int64) {
    reader := bufio.NewReader(input)

    for {
        line, err := reader.ReadString('\n')
        fields := strings.Split(line, delimiter)

        collectedFields := collectFields(fields, selectedFields)

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

func main() {
    arguments, err := parseArguments(os.Args[1:])
    if err != nil {
        fmt.Println("Invalid arguments:", err)
        return
    }

    for _, file := range arguments.files {
        cut(file, os.Stdout, arguments.delimiter, arguments.fields)
    }

}
