package main

import (
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
)

func writeMemoryProfile() string {
	f, err := ioutil.TempFile("/tmp", "frank_memory_debug_")
	if err != nil {
		log.Printf("Failed to write memory debugging file: %v", err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
	os.Chmod(f.Name(), 0644)
	log.Printf("Saved memory profile to %s", f.Name())

	return f.Name()
}
