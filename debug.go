package main

import (
	"log"
	"os"
	"runtime/pprof"
)

func writeMemoryProfile() {
	if *memprofile == "" {
		return
	}

	log.Printf("Saving memory profile to %s", *memprofile)
	f, err := os.Create(*memprofile)
	if err != nil {
		log.Fatal(err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
}
