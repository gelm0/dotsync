package dotsync

import (
	"errors"
	"io"
	"os"

	diff "github.com/sergi/go-diff/diffmatchpatch"
)

// Diffs two files by reading them in chunks of 1024 bytes
// first checks if the two files are of different size and if no
// difference is found continues to check the difference
// returns true if difference is found
func DiffFiles(filePath1 string, filePath2 string) (bool, error) {
	if filePath1 == "" || filePath2 == "" {
		return true, errors.New("Empty path supplied")
	}
	file1, err := os.Open(filePath1)
	if err != nil {
		// Might change this to log directly later, for now we return it
		return true, err
	}
	file2, err := os.Open(filePath2)
	if err != nil {
		// Might change this to log directly later, for now we return it
		return true, err
	}
	// Make sure we close files when we are done
	defer file1.Close()
	defer file2.Close()
	// First check if there is a difference in file size
	sizeDiffers, err := diffSize(file1, file2)
	if sizeDiffers || err != nil {
		return sizeDiffers, err
	}
	// We did not find any file size difference but we can not be sure that nothing changed
	return diffChars(file1, file2)
}

// Diffs size of two files, returns true if difference is found
func diffSize(file1, file2 *os.File) (bool, error) {
	f1Info, err := file1.Stat()
	if err != nil {
		return true, err
	}
	f2Info, err := file1.Stat()
	if err != nil {
		return true, err
	}
	return f1Info.Size() != f2Info.Size(), nil
}

// // Start by diffing first 1024 and last 1024 bytes
func diffChars(file1, file2 *os.File) (bool, error) {
	const bufferSize = 1024
	buffer1 := make([]byte, bufferSize)
	buffer2 := make([]byte, bufferSize)
	diffMatchPatch := diff.New()
	// We don't want to break the loop until the diff has been run
	breakLoop := false
	for {
		read1, err := file1.Read(buffer1)
		if err != nil {
			if err != io.EOF {
				return true, err
			}
			breakLoop = true
		}
		read2, err := file2.Read(buffer2)
		if err != nil {
			if err != io.EOF {
				return true, err
			}
			breakLoop = true
		}
		difference := diffMatchPatch.DiffMain(string(buffer1[:read1]), string(buffer2[:read2]), false)
		if len(difference) > 0 {
			for i := 0; i < len(difference); i++ {
				// 0 is equivalent to Equals, hence we check that everything that is not equal
				// is returned as a difference
				if difference[i].Type != 0 {
					return true, nil
				}
			}
		}
		if breakLoop {
			break
		}
	}
	return false, nil
}
