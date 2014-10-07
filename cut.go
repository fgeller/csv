package main

import "flag"
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
    var flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

    var fields = flags.String("f", "", fields_message)
    var delimiter = flags.String("d", ",", delimiter_message)

    flags.Parse(arguments)

    return map[string]interface{}{
        "fields":    *fields,
        "delimiter": *delimiter,
    }
}

func delimiter(arguments map[string]interface{}) string {
    return arguments["delimiter"].(string)
}

func selectedFields(arguments map[string]interface{}) []int64 {
    selectedFields := arguments["fields"].(string)
    if 0 == len(selectedFields) { // TODO: why is 1 = len([]string{})
        return []int64{}
    }

    splits := strings.Split(selectedFields, ",")
    numbers := make([]int64, len(splits))
    for idx, stringNumber := range splits {
        numbers[idx], _ = strconv.ParseInt(stringNumber, 10, 64)
    }

    return numbers
}

func cut(input io.Reader, output io.Writer, delimiter string, selectedFields []int64) {

    reader := bufio.NewReader(input)

    for {
        line, err := reader.ReadString('\n')
        fields := strings.Split(line, delimiter)
        collectedFields := make([]string, len(selectedFields))

        for index, fieldIndex := range selectedFields {
            collectedFields[index] = fields[fieldIndex-1]
        }

        newLine := fmt.Sprintln(strings.TrimSuffix(strings.Join(collectedFields, delimiter), "\n"))
        io.WriteString(output, newLine)

        if err == io.EOF {
            break
        }

        if err != nil {
            fmt.Println("Encountered error while reading:", err)
            break
        }
    }
}

func main() {
    // delimiter := delimiter(arguments)
    // selectedFields := selectedFields(arguments)
    // arguments := parseArguments(os.Args[1:])
}
