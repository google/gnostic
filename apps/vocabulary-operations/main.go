// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	metrics "github.com/googleapis/gnostic/metrics"
	vocabulary "github.com/googleapis/gnostic/metrics/vocabulary"
)

// openVocabularyFiles uses standard input to create a slice of
// Vocabulary protocol buffer filenames.
// The slice of filenames is returned and will be used to createe
// Vocabulary structures.
func openVocabularyFiles() []string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)
	var v []string
	for scanner.Scan() {
		v = append(v, scanner.Text())
	}
	return v
}

// readVocabularyFromFilename accepts the filename of a Vocabulary pb
// and parses the data in the file which is then added to a Vocabulary struct.
func readVocabularyFromFilename(filename string) *metrics.Vocabulary {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File %s error: %v\n", filename, err)
		os.Exit(1)
	}

	v := &metrics.Vocabulary{}
	err = proto.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
	return v
}

// processInputs determines whether the application will be using cmd line args or stdin.
// The function takes the cmd lines arguments, if any, and a flag which is true if stdin
// will be the source of input.
func processInputs(args []string, stdinFlag bool) []*metrics.Vocabulary {
	v := make([]*metrics.Vocabulary, 0)
	switch stdinFlag {
	case true:
		files := openVocabularyFiles()
		for _, file := range files {
			v = append(v, readVocabularyFromFilename(file))
		}
		return v
	default:
		files := make([]string, 0)
		for _, arg := range args {
			files = append(files, arg)
		}
		for _, file := range files {
			v = append(v, readVocabularyFromFilename(file))
		}
		return v
	}
}

func main() {
	unionPtr := flag.Bool("union", false, "generates the union of pb files")
	intersectionPtr := flag.Bool("intersection", false, "generates the intersection of pb files")
	differencePtr := flag.Bool("difference", false, "generates the difference of pb files")
	exportPtr := flag.Bool("export", false, "export a given pb file as a csv file")
	filterCommonPtr := flag.Bool("filter-common", false, "egenerates uniqueness within company")

	flag.Parse()
	args := flag.Args()
	if !*unionPtr && !*intersectionPtr && !*differencePtr && !*exportPtr && !*filterCommonPtr {
		flag.PrintDefaults()
		fmt.Printf("Please use one of the above command line arguments.\n")
		os.Exit(-1)
		return
	}
	vocabularies := make([]*metrics.Vocabulary, 0)
	switch arguments := len(args); arguments {
	case 0:
		vocabularies = processInputs(args, true)
	default:
		vocabularies = processInputs(args, false)
	}

	if *unionPtr {
		vocab := vocabulary.Union(vocabularies)
		vocabulary.WritePb(vocab)
	}
	if *intersectionPtr {
		vocab := vocabulary.Intersection(vocabularies)
		vocabulary.WritePb(vocab)
	}
	if *differencePtr {
		vocab := vocabulary.Difference(vocabularies)
		vocabulary.WritePb(vocab)
	}
	if *exportPtr {
		vocabulary.WriteCSV(vocabularies[0], "")
	}
	if *filterCommonPtr {
		vocab := vocabulary.FilterCommon(vocabularies)
		vocabulary.WritePb(vocab[0])
	}
}
