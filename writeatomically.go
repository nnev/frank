package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func tempDir(dest string) string {
	tempdir := os.Getenv("TMPDIR")
	if tempdir == "" {
		// Convenient for development: decreases the chance that we
		// cannot move files due to TMPDIR being on a different file
		// system than dest.
		tempdir = filepath.Dir(dest)
	}
	return tempdir
}

func writeAtomically(dest string, write func(w io.Writer) error) (err error) {
	f, err := ioutil.TempFile(tempDir(dest), "atomic-")
	if err != nil {
		return err
	}
	defer func() {
		// Clean up (best effort) in case we are returning with an error:
		if err != nil {
			// Prevent file descriptor leaks.
			f.Close()
			// Remove the tempfile to avoid filling up the file system.
			os.Remove(f.Name())
		}
	}()

	// Use a buffered writer to minimize write(2) syscalls.
	bufw := bufio.NewWriter(f)

	if err := write(bufw); err != nil {
		return err
	}

	if err := bufw.Flush(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return os.Rename(f.Name(), dest)
}
