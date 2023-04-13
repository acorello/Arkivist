package osutil

import (
	"log"
	"os"
)

func MustOpen(filepath string) *os.File {
	r, err := os.Open(filepath)
	if err != nil {
		log.Fatal()
	}
	return r
}
