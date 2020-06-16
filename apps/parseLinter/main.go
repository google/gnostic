package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	ibmPtr := flag.Bool("IBM", false, "generates the linter proto for IBM outputs")
	spectralPtr := flag.Bool("Spectral", false, "generates the linter proto for Spectral outputs")

	flag.Parse()
	args := flag.Args()

	if !*ibmPtr && !*spectralPtr {
		flag.PrintDefaults()
		fmt.Printf("Please use one of the above command line arguments.\n")
		os.Exit(-1)
		return
	}

	if len(args) != 1 {
		fmt.Printf("Usage: report <file.json>\n")
		return
	}

	if *ibmPtr {
		lintIBM(args[0])
	}

	if *spectralPtr {
		lintSpectral(args[0])
	}

}
