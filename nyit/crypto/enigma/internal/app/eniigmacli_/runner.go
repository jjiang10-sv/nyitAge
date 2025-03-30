package eniigmacli_

import (
	"bufio"
	"errors"
	"os"
)

func Runner() error {
	info, err := createParseInfo(os.Args, os.Stdout)
	if err != nil {
		return errors.New("can not create information from commandline args")
	}
	if info.isHelp {
		return nil
	}

	fileoutput := os.Stdout
	if info.fileName != "" {
		fileoutput, err = os.Create(info.fileName)
		if err != nil {
			return nil
		}
		defer fileoutput.Close()
		fileoutput := bufio.NewWriter(fileoutput)
		defer fileoutput.Flush()

	}
	return runNormalMode(info, os.Stdin, os.Stdout, fileoutput)
}
