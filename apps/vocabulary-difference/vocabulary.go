package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	metrics "github.com/googleapis/gnostic/metrics"
	"google.golang.org/protobuf/proto"
)

func WriteCSV(v *metrics.Vocabulary) {
	f4, ferror := os.Create("summarize-vocabulary.csv")
	if ferror != nil {
		fmt.Println(ferror)
		f4.Close()
		return
	}
	for _, s := range v.Schemas {
		temp := fmt.Sprintf("%s,\"%s\",%d\n", "schemas", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Properties {
		temp := fmt.Sprintf("%s,\"%s\",%d\n", "properties", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Operations {
		temp := fmt.Sprintf("%s,\"%s\",%d\n", "operations", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	for _, s := range v.Parameters {
		temp := fmt.Sprintf("%s,\"%s\",%d\n", "parameters", s.Word, int(s.Count))
		f4.WriteString(temp)
	}
	f4.Close()
}

func WritePb(v *metrics.Vocabulary) {
	bytes, err := proto.Marshal(v)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("summarize-vocabulary.pb", bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func FillProtoStructures(m map[string]int) []*metrics.WordCount {
	key_names := make([]string, 0, len(m))
	for key := range m {
		key_names = append(key_names, key)
	}
	sort.Strings(key_names)

	counts := make([]*metrics.WordCount, 0)
	for _, k := range key_names {
		temp := &metrics.WordCount{
			Word:  k,
			Count: int32(m[k]),
		}
		counts = append(counts, temp)
	}
	return counts
}

func UnpackageVocabulary(v *metrics.Vocabulary) {
	for _, s := range v.Schemas {
		schemas[s.Word] += int(s.Count)
	}
	for _, op := range v.Operations {
		operationID[op.Word] += int(op.Count)
	}
	for _, param := range v.Parameters {
		parameters[param.Word] += int(param.Count)
	}
	for _, prop := range v.Properties {
		properties[prop.Word] += int(prop.Count)
	}
}