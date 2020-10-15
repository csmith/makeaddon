package main

import (
	"github.com/csmith/makeaddon"
	"log"
	"os"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Unable to get working directory: %v", err)
	}

	f, err := os.Create("addon.zip")
	if err != nil {
		log.Fatalf("Unable to create output file: %v", err)
	}

	builder, err := makeaddon.NewBuilder(cwd, f)
	if err != nil {
		log.Fatalf("Unable to create builder: %v", err)
	}

	err = builder.Build()
	if err != nil {
		log.Fatalf("Unable to build: %v", err)
	}

	f.Close()
}
