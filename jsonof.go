package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
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

	if s == "null" || s == "nil" {
		return nil
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return i
	}

	f, err := strconv.ParseFloat(s, 10)
	if err == nil {
		return f
	}

	b, err := strconv.ParseBool(s)
	if err == nil {
		return b
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
