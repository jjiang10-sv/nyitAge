package main

import (
	"fmt"
	"os"

	"github.com/ibraimgm/enigma/internal/app/eniigmacli_"
)

func main() {
	if err := eniigmacli_.Runner(); err != nil {
		fmt.Fprintf(os.Stderr, "***Error %v \n", err)
		os.Exit(1)
	}
}
