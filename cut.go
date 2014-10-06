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

    scanner := bufio.NewScanner(input)
    for scanner.Scan() {
        truncatedLine := scanner.Text()
        io.WriteString(output, fmt.Sprintln(truncatedLine))
    }

    if err := scanner.Err(); err != nil {
        fmt.Println(os.Stderr, "Err while reading from input:", err)
    }
}

func main() {
    // arguments := parseArguments(os.Args[1:])
}
