// Tomll is a linter for TOML
//
// Usage:
//   cat file.toml | tomll > file_linted.toml
//   tomll file1.toml file2.toml # lint the two files in place
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/fletaio/fleta/cmd/config/go-toml"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "tomll can be used in two ways:")
		fmt.Fprintln(os.Stderr, "Writing to STDIN and reading from STDOUT:")
		fmt.Fprintln(os.Stderr, "  cat file.toml | tomll > file.toml")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Reading and updating a list of files:")
		fmt.Fprintln(os.Stderr, "  tomll a.toml b.toml c.toml")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "When given a list of files, tomll will modify all files in place without asking.")
	}
	flag.Parse()
	// read from stdin and print to stdout
	if flag.NArg() == 0 {
		s, err := lintReader(os.Stdin)
		if err != nil {
			io.WriteString(os.Stderr, err.Error())
			os.Exit(-1)
		}
		io.WriteString(os.Stdout, s)
	} else {
		// otherwise modify a list of files
		for _, filename := range flag.Args() {
			s, err := lintFile(filename)
			if err != nil {
				io.WriteString(os.Stderr, err.Error())
				os.Exit(-1)
			}
			ioutil.WriteFile(filename, []byte(s), 0644)
		}
	}
}

func lintFile(filename string) (string, error) {
	tree, err := toml.LoadFile(filename)
	if err != nil {
		return "", err
	}
	return tree.String(), nil
}

func lintReader(r io.Reader) (string, error) {
	tree, err := toml.LoadReader(r)
	if err != nil {
		return "", err
	}
	return tree.String(), nil
}
