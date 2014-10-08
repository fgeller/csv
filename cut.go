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

func parseArguments(arguments []string) map[string]interface{} {
    // default values
    fields := ""
    delimiter := ","
    fileNames := []string{}

    for index := 0; index < len(arguments); index += 1 {
        argument := arguments[index]
        switch {
        case argument == "-f":
            fields = arguments[index+1]
            index += 1
        case strings.HasPrefix(argument, "-f"):
            fields = argument[2:]
        case argument == "-d":
            delimiter = arguments[index+1]
            index += 1
        case strings.HasPrefix(argument, "-d"):
            delimiter = argument[2:]
        case true:
            fileNames = append(fileNames, argument)
        }
    }

    return map[string]interface{}{
        "fields":    fields,
        "delimiter": delimiter,
        "fileNames": fileNames,
    }
}

func files(arguments map[string]interface{}) ([]*os.File, error) {
    fileNames := arguments["fileNames"].([]string)
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

func delimiter(arguments map[string]interface{}) string {
    return arguments["delimiter"].(string)
}

func selectedFields(arguments map[string]interface{}) []int64 {
    selectedFields := arguments["fields"].(string)
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
    arguments := parseArguments(os.Args[1:])
    files, err := files(arguments)
    if err != nil {
        fmt.Println("Encountered error while opening files:", err)
        return
    }

    for _, file := range files {
        cut(file, os.Stdout, delimiter(arguments), selectedFields(arguments))
    }

}
