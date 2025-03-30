package eniigmacli_

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ibraimgm/enigma/machine/enigma1"
	"github.com/ibraimgm/enigma/machine/parts1"
	"github.com/pborman/getopt/v2"
)

type parseInfo struct {
	e         enigma1.Enigma
	isHelp    bool
	isQuiet   bool
	fileName  string
	blockSize uint
}

func createParseInfo(args []string, stdout io.Writer) (*parseInfo, error) {
	//getopt.CommandLine = getopt.New()
	isHelp := getopt.BoolLong("help", 'h', "to show the help")
	isQuite := getopt.BoolLong("quite", 'q', "isQuite flag not print out the logs")
	rotorsOpt := getopt.StringLong("rotors", 'r', "III,II,I", "Comma-separated list of rotors to be used.", "III,II,I")
	reflectorOpt := getopt.StringLong("reflector", 'f', "B", "Reflector to use.", "B")
	ringOpt := getopt.StringLong("ring", 'g', "AAA", "Ring settings to be used.", "ABC")
	windowOpt := getopt.StringLong("window", 'w', "AAA", "Window settings to be used.", "ABC")
	blockOpt := getopt.IntLong("blocksize", 'b', 5, "Block size of the coded text (default: 5)")
	fileOpt := getopt.StringLong("output", 'o', "", "Output file to write.", "a.txt")

	// if err := parseGetopt(args); err != nil {
	// 	return nil, err
	// }

	if *isHelp {
		getopt.PrintUsage(stdout)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "All command-line arguments are optional.")
		fmt.Fprintln(stdout, "By default, enigma run in 'normal' mode, which reads one line from sdtin and outputs encoded text, until EOF is reached.")
		fmt.Fprintln(stdout, "This means that after writing a line and pressing 'Enter', the coded version will be displayed immediately (written to file).")
		fmt.Fprintln(stdout, "The coding process will output the characters in 'blocks', whose size can be controlled with the '-b' flag.")
		return &parseInfo{isHelp: true}, nil
	}
	if _, ok := parts1.Reflectors[*reflectorOpt]; !ok {
		return nil, errors.New("can not get reflector")
	}
	rotorIds := strings.Split(*rotorsOpt, ",")
	if len(rotorIds) != 3 {
		return nil, errors.New("only three rotors are allowed here")
	}
	e, err := enigma1.WithRotors(rotorIds[0], rotorIds[1], rotorIds[2], *reflectorOpt)
	if err != nil {
		return nil, err
	}
	e.Configure(*windowOpt, *ringOpt)
	//return &parseInfo{e, *fileOpt, *quietOpt, false, uint(*blockOpt)}, nil
	return &parseInfo{e, *isHelp, *isQuite, *fileOpt, uint(*blockOpt)}, nil

}

func parseGetopt(args []string) error {
	oldArgs := os.Args
	os.Args = args
	defer func() { os.Args = oldArgs }()
	return getopt.Getopt(nil)
}
