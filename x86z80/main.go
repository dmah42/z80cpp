package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	in = flag.String("in", "", "the file to read")
	labelRE = regexp.MustCompile(`^\.?(\w+):`)
	opRE = regexp.MustCompile(`^\s+(\w+)\s+([\w,\.\d\s\[\]]+)`)
)

func main() {
	flag.Parse()

	log.Printf("opening %q", *in)
	file, err := os.Open(*in)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var trans string
		if labelRE.MatchString(line) {
			label := labelRE.FindStringSubmatch(line)[1]
			trans = fmt.Sprintf("LABEL: %q", label)
		} else if opRE.MatchString(line) {
			op := opRE.FindStringSubmatch(line)
			operator := op[1]
			operands := strings.Split(op[2], ",")
			for i, o := range operands {
				operands[i] = strings.TrimSpace(o)
			}
			trans = fmt.Sprintf("OP: %q %q", operator, operands)
		}

		fmt.Printf("%q -> %s\n", line, trans)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
