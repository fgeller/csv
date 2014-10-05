package main

import "flag"
import "os"
import "fmt"
import "bufio"
import "io"

const (
    fields_message string = "select only these fields"
)

func parseArguments(arguments []string) map[string]interface{} {
    var flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    var fields = flags.String("f", "default", fields_message)

    flags.Parse(arguments)

    return map[string]interface{}{
        "fields": *fields,
    }
}

func cut(input io.Reader, output io.Writer, arguments map[string]interface{}) {
    bufInput := bufio.NewReader(input)

    firstLine, _, _ := bufInput.ReadLine()
    fmt.Println("read line:", string(firstLine))

    io.WriteString(output, string(firstLine))
}

func main() {
    arguments := parseArguments(os.Args[1:])
    fmt.Println("arguments", arguments)
}
