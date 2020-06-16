package main

import(
	"os"
	"fmt"
	"log"
	"bufio"
	"regexp"
	"strconv"

	linter "github.com/googleapis/gnostic/metrics/linterResult"
)

func parseOutput(output []string) []*linter.Message{
	messages := make([]*linter.Message, 0)
	for _, line := range output{
		array := regexp.MustCompile("[]: *]").Split(line, 6)
		line,_ := strconv.ParseInt(array[1], 0, 64)
		temp := &linter.Message{
			Type:    array[3],
			Message: array[5],
			Line:    int32(line),
		}
		messages = append(messages,temp)
	}
	return messages
}

func openAndReadTxt(filename string) []string{
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	output := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return output
}

func lintSpectral(filename string) {
	output := openAndReadTxt(filename)
	messages := parseOutput(output)
	linterResult := &linter.Linter{
		LinterResults: messages,
	}
	fmt.Printf("%+v", linterResult)
	writePb(linterResult)
}