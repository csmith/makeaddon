package main

import (
	"flag"
	"github.com/csmith/makeaddon"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Unable to get working directory: %v", err)
	}

	f, err := os.Create("addon.zip")
	if err != nil {
		log.Fatalf("Unable to create output file: %v", err)
	}

	builder, err := makeaddon.NewBuilder(cwd, filepath.Base(cwd), f)
	if err != nil {
		_ = os.Remove("addon.zip")
		log.Fatalf("Unable to create builder: %v", err)
	}

	err = builder.Build()
	if err != nil {
		_ = os.Remove("addon.zip")
		log.Fatalf("Unable to build: %v", err)
	}

	f.Close()
}
