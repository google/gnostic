package lint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"google.golang.org/protobuf/proto"
)

type Lint struct {
	LinterErrors   ErrorResult   `json:"errors"`
	LinterWarnings WarningResult `json:"warnings"`
}

type ErrorResult struct {
	Parameters []EMessage `json:"parameters-ibm"`
	PathsIBM   []EMessage `json:"paths-ibm"`
	Paths      []WMessage `json:"paths"`
	Schemas    []EMessage `json:"schema-ibm"`
	FormData   []WMessage `json:"form-data"`
	WalkerIBM  []EMessage `json:"walker-ibm"`
}

type WarningResult struct {
	OperationID      []WMessage `json:"operation-ids"`
	Operations       []WMessage `json:"operation"`
	OperationsShared []WMessage `json:"operations-shared"`
	Refs             []WMessage `json:"refs"`
	Schemas          []EMessage `json:"schema-ibm"`
	PathsIBM         []EMessage `json:"paths-ibm"`
	WalkerIBM        []EMessage `json:"walker-ibm"`
	CircularIBM      []WMessage `json:"circular-references-ibm"`
	Responses        []EMessage `json:"responses"`
	ParametersIBM    []EMessage `json:"parameters-ibm"`
}

type EMessage struct {
	Path    []string `json:"path"`
	Message string   `json:"message"`
	Line    int      `json:"line"`
}

type WMessage struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Line    int    `json:"line"`
}

func writePb(v *Linter) {
	bytes, err := proto.Marshal(v)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("linterResults.pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func addToMessages(mtype string, message string, path string, line int) *Message {
	temp := &Message{
		Type:    mtype,
		Message: message,
		Path:    path,
		Line:    int32(line),
	}
	return temp
}

func fillMessageProtoStructureIBM(lint Lint) []*Message {
	messages := make([]*Message, 0)
	for _, v := range lint.LinterErrors.Parameters {
		temp := addToMessages("Error", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterErrors.PathsIBM {
		temp := addToMessages("Error", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterErrors.Paths {
		temp := addToMessages("Error", v.Message, v.Path, v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterErrors.Schemas {
		temp := addToMessages("Error", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterErrors.FormData {
		temp := addToMessages("Error", v.Message, v.Path, v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterErrors.WalkerIBM {
		temp := addToMessages("Error", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.OperationID {
		temp := addToMessages("Warning", v.Message, v.Path, v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.OperationsShared {
		temp := addToMessages("Warning", v.Message, v.Path, v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.Refs {
		temp := addToMessages("Warning", v.Message, v.Path, v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.Schemas {
		temp := addToMessages("Warning", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.PathsIBM {
		temp := addToMessages("Warning", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.WalkerIBM {
		temp := addToMessages("Warning", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.CircularIBM {
		temp := addToMessages("Warning", v.Message, v.Path, v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.Operations {
		temp := addToMessages("Warning", v.Message, v.Path, v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.Responses {
		temp := addToMessages("Warning", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}
	for _, v := range lint.LinterWarnings.ParametersIBM {
		temp := addToMessages("Warning", v.Message, strings.Join(v.Path, "."), v.Line)
		messages = append(messages, temp)
	}

	return messages
}

func openAndReadJSON(filename string) Lint {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var lint Lint

	json.Unmarshal(byteValue, &lint)

	return lint
}

func LintIBM(filename string) {
	lint := openAndReadJSON(filename)
	messages := fillMessageProtoStructureIBM(lint)

	linterResult := &Linter{
		LinterResults: messages,
	}

	writePb(linterResult)
}
