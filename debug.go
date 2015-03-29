package main

import (
	"io/ioutil"
	"log"
	"runtime/pprof"
)

func writeMemoryProfile() string {
	f, err := ioutil.TempFile("/tmp", "frank_memory_debug_")
	if err != nil {
		log.Printf("Failed to write memory debugging file: %v", err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
	log.Printf("Saved memory profile to %s", f.Name())

	return f.Name()
}
