package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// global settings
	OneLineOutput bool
	NoBreaking    bool
)

func main() {

	flag.BoolVar(&OneLineOutput, "noindent", false, "single line output, don't pretty print")
	flag.BoolVar(&NoBreaking, "nobreak", false, `do not pre-process input values (e.g. don't do "[{ x: y }]" -> "[ { x : y } ]"`)

	flag.Parse()

	values := flag.Args()
	if !NoBreaking {
		values = breakUpInput(values)
	}

	// log.Printf("values: %s", strings.Join(values, ", "))

	data := maybeArgsToArray(values)

	os.Stdout.Write(marshal(data))
	os.Stdout.WriteString("\n")
}

func marshal(data any) []byte {
	var output []byte
	var err error

	if OneLineOutput {
		output, err = json.Marshal(data)
	} else {
		output, err = json.MarshalIndent(data, "", "    ")
	}
	if err != nil {
		log.Fatalf("failed to marshal as JSON: %v", err)
	}

	return output
}

func maybeArgsToArray(args []string) any {

	values, remain := argsToArray(args)

	if len(remain) > 0 {
		log.Printf("Warning: remaining input not processed: %v", strings.Join(remain, " "))
	}

	if len(values) == 1 {
		return values[0]
	}

	return values
}

// argsToArray processes args collecting json values into an "array" (a slice of any),
// finishing at "]" or the end of the args. Returns the collected items, and any remaining
// strings left from the input args.
func argsToArray(args []string) ([]any, []string) {

	result := []any{}

	for i := 0; i < len(args); i++ {

		switch args[i] {
		case "[":
			arr, remain := argsToArray(args[i+1:])
			result = append(result, arr)
			args = remain
			i = -1
		case "{":
			obj, remain := argsToObject(args[i+1:])
			result = append(result, obj)
			args = remain
			i = -1
		case "]":
			return result, args[i+1:]
		case "}":
			return result, args[i:]
		default:
			result = append(result, valOf(args[i]))
		}
	}

	return result, []string{}
}

// argsToObject processes args collecting name/value pairs into a map, until it
// finds a "}" or "]" or the end of the args. Returns the map, and any remaining
// strings left over.
func argsToObject(args []string) (map[string]any, []string) {

	result := map[string]any{}

	name := ""
	for i := 0; i < len(args); i++ {

		switch args[i] {
		case "[":
			arr, remain := argsToArray(args[i+1:])
			result[nameOrMissing(name)] = arr
			name = ""
			args = remain
			i = -1
		case "{":
			obj, remain := argsToObject(args[i+1:])
			result[nameOrMissing(name)] = obj
			name = ""
			args = remain
			i = -1
		case "}":
			if name != "" {
				result[name] = ""
			}
			return result, args[i+1:]
		case "]":
			if name != "" {
				result[name] = ""
			}
			return result, args[i:]
		case ":":
			name = nameOrMissing(name)
		default:
			if name == "" {
				name = args[i]
			} else {
				result[name] = valOf(args[i])
				name = ""
			}
		}
	}

	if name != "" {
		result[name] = ""
	}

	return result, []string{}
}

// nameOrMissing returns the input if not blank, otherwise returns "*missing-name*".
func nameOrMissing(name string) string {
	if name == "" {
		return "*missing-name*"
	}
	return name
}

// breakUpInput will break up any passed arguments that contain JSON-special characters ({, }, [, ], or :).
func breakUpInput(values []string) []string {

	// guestimate a size for result slice.
	size := len(values) + (len(values) / 10)
	result := make([]string, 0, size)

	i := 0
	for i < len(values) {

		v := values[i]

		p := strings.IndexAny(v, "{}[]:")
		if p == -1 {
			result = append(result, v)
			i++
			continue
		}

		if p > 0 {
			result = append(result, v[0:p])
		}
		result = append(result, string(v[p]))

		remain := v[p+1:]
		if len(remain) == 0 {
			i++
			continue
		}

		values[i] = remain
	}

	return result
}

func run(args []string, stdout io.Writer) error {

	output := map[string]interface{}{}
	key := ""
	values := []interface{}{}
	pretty := false

	for _, arg := range args[1:] {
		if arg == "-p" || arg == "--pretty" {
			pretty = true
			continue
		}

		if strings.HasSuffix(arg, ":") {
			output = collectOutput(output, key, values)
			key = strings.TrimSuffix(arg, ":")
			values = values[:0]
			continue
		}

		values = append(values, arg)
	}

	output = collectOutput(output, key, values)

	switch pretty {
	case true:
		fmt.Println(mustString(json.MarshalIndent(output, "", "    ")))
	case false:
		fmt.Println(mustString(json.Marshal(output)))
	}

	return nil
}

func collectOutput(output map[string]interface{}, key string, values []interface{}) map[string]interface{} {

	if len(values) == 0 {
		return output
	}

	if key == "" {
		log.Fatalf("value(s) \"%s\" without key: do jsonobj key: val key: val ...", values)
	}

	if len(values) == 1 {
		output[key] = valOf(values[0])
	} else {
		output[key] = valOfs(values...)
	}

	return output
}

func mustString(content []byte, err error) string {
	if err != nil {
		log.Fatalf("%v", err)
	}

	return string(content)
}

func valOfs(input ...interface{}) []interface{} {

	res := []interface{}{}
	for _, v := range input {
		res = append(res, valOf(v))
	}

	return res
}

func valOf(input interface{}) interface{} {

	s, ok := input.(string)
	if !ok {
		return input
	}

	switch s {
	case "null", "nil":
		return nil
	case "true", "TRUE", "True":
		return true
	case "false", "FALSE", "False":
		return false
	case "@nowlocal":
		return now()
	case "@now":
		return nowUTC()
	case "@uuid":
		return newUUID()
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return i
	}

	f, err := strconv.ParseFloat(s, 10)
	if err == nil {
		return f
	}

	// default string
	return s
}

func newUUID() string {
	return uuid.New().String()
}

func now() string {
	t := time.Now()
	return t.Format("2006-01-02T15:04:05Z07:00")
}

func nowUTC() string {
	t := time.Now().UTC()
	return t.Format("2006-01-02T15:04:05Z07:00")
}
