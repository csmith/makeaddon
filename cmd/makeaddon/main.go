package main

import (
	"flag"
	"github.com/csmith/makeaddon"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	flag.Parse()

	var target string
	var name string

	firstArg := flag.Arg(0)
	if firstArg == "" || firstArg == "." {
		// No arguments or just passing '.' - use the current working dir
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Unable to get working directory: %v", err)
		}
		target = cwd
		name = filepath.Base(cwd)
	} else if stat, err := os.Stat(firstArg); err == nil && stat.IsDir() {
		// An argument that appears to be a local directory - use that!
		target = firstArg
		name = filepath.Base(firstArg)
	} else {
		// Assume it's a VCS url
		dir, err := makeaddon.Checkout(firstArg, flag.Arg(1))
		if err != nil {
			log.Fatalf("Unable to check out addon from VCS: %v", err)
		}
		target = dir
		name = strings.TrimSuffix(path.Base(firstArg), ".git")
	}

	f, err := os.Create("addon.zip")
	if err != nil {
		log.Fatalf("Unable to create output file: %v", err)
	}

	builder, err := makeaddon.NewBuilder(target, name, f)
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
