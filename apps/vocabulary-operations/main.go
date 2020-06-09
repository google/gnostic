package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	metrics "github.com/googleapis/gnostic/metrics"
)

/*
These variables were made globally because multiple
functions will be accessing and mutating them.
*/
var schemas map[string]int
var operationID map[string]int
var parameters map[string]int
var properties map[string]int

func openVocabularyFiles() []string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	var v []string
	for scanner.Scan() {
		v = append(v, scanner.Text())
	}
	return v
}

func readVocabularyFromFilename(filename string) *metrics.Vocabulary {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	v := &metrics.Vocabulary{}
	err = proto.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
	return v
}

func main() {
	unionPtr := flag.Bool("union", false, "generates the union of pb files")
	intersectionPtr := flag.Bool("intersection", false, "generates the intersection of pb files")
	differencePtr := flag.Bool("difference", false, "generates the difference of pb files")
	exportPtr := flag.Bool("export", false, "export a given pb file as a csv file")

	flag.Parse()
	args := flag.Args()

	if !*unionPtr && !*intersectionPtr && !*differencePtr && !*exportPtr {
		flag.PrintDefaults()
		fmt.Printf("Please use one of the above command line arguments.\n")
		os.Exit(-1)
		return
	}
	v := make([]*metrics.Vocabulary, 0)
	switch arguments := len(args); arguments {
	case 0:
		files := openVocabularyFiles()
		for _, file := range files {
			v = append(v, readVocabularyFromFilename(file))
		}
	default:
		files := make([]string, 0)
		for _, arg := range args {
			files = append(files, arg)
		}
		for _, file := range files {
			v = append(v, readVocabularyFromFilename(file))
		}
	}

	if *unionPtr {
		writePb(vocabularyUnion(v))
	}
	if *intersectionPtr {
		writePb(vocabularyIntersection(v))
	}
	if *differencePtr {
		writePb(vocabularyDifference(v))
	}
	if *exportPtr {
		writeCSV(v[0])
	}
}
