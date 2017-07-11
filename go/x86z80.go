package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	in = flag.String("in", "test.asm", "the file to read")

	labelRE = regexp.MustCompile(`^\.?(\w+):`)
	opRE = regexp.MustCompile(`^(\w+)\s+(.*)`)

	// TODO: translate registers.
	trMap = map[string]func(...string) []string {
		"MOV": func(args ...string) []string {
			return []string {
				// Z80: LD loads the second arg into the first.
				// Either arg can be a register, an 8bit integer, (HL) (the contents of register HL),
				// or (IX + d) or (IY + d) where d is a two's complement integer offset.
				// TODO: parse the args figure out which modes to use.
				fmt.Sprintf("LD a, %s", args[1]),
				fmt.Sprintf("LD %s, a", args[0]),
			}
		},
		"AND": func(args ...string) []string {
			return []string {
				// Z80: arg can be a register, a constant (8-bit), (HL), (IX+d), or (IY+d), d being 8-bit.
				fmt.Sprintf("LD a, %s", args[0]),
				fmt.Sprintf("AND %s", args[1]),
				fmt.Sprintf("LD %s, a", args[0]),
			}
		},
		"SHL": func(args ...string) []string {
			return []string {
				// Z80: arg can be the usual set.
				fmt.Sprintf("LD a, %s", args[0]),
				fmt.Sprintf("SLA a"),
				fmt.Sprintf("LD %s, a", args[0]),
			}
		},
		"SHR": func(args ...string) []string {
			return []string {
				// Z80: arg can be the usual set.
				fmt.Sprintf("LD a, %s", args[0]),
				fmt.Sprintf("SRA a"),
				fmt.Sprintf("LD %s, a", args[0]),
			}
		},
		"JMP": func(args ...string) []string {
			return []string {
				// Z80: JP takes a 16-bit integer address or (HL), (IX), or (IY), unconditionally.
				fmt.Sprintf("JP %s", args[0]),
			}
		},
	}
)

func stripComments(line string) string {
	return strings.TrimSpace(strings.Split(line, "#")[0])
}

func formatLabel(label string) string {
	return fmt.Sprintf("%s:\n", label)
}

func formatOp(operator string, operands []string) (string, error) {
	if op, ok := trMap[operator]; ok {
		ops := op(operands...)
		var trans string
		if len(ops) != 0 {
			trans = fmt.Sprintf("        %s\n", ops[0])
			for _, o := range ops[1:] {
				trans += fmt.Sprintf("        %s\n", o)
			}
		}
		return trans, nil
	}
	return "", fmt.Errorf("operator not found in mapping: %q", operator)
}

func translate(line string) (string, error) {
	line = stripComments(line)
	maybeLabel := labelRE.FindStringSubmatch(line)
	if maybeLabel != nil {
		return formatLabel(maybeLabel[1]), nil
	}

	maybeOp := opRE.FindStringSubmatch(line)
	if maybeOp != nil {
		operator := strings.ToUpper(maybeOp[1])
		operands := strings.Split(maybeOp[2], ",")
		for i, o := range operands {
			operands[i] = strings.TrimSpace(o)
		}
		return formatOp(operator, operands)
	}
	return "", fmt.Errorf("line does not match regexp: %q", line)
}

func main() {
	flag.Parse()

	absin, err := filepath.Abs(*in)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("opening %q for reading", absin)
	infile, err := os.Open(absin)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	out := fmt.Sprintf("%s.z80", strings.TrimSuffix(absin, path.Ext(*in)))
	log.Printf("opening %q for writing", out)
	outfile, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	fmt.Fprintf(outfile, "ORG 32768\n")

	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprint(outfile, fmt.Sprintf("# %s\n", line))
		trans, err := translate(line)
		if err != nil {
			log.Printf(err.Error())
			trans = fmt.Sprintf("# ERROR: %s\n", err)
		}
		fmt.Fprint(outfile, trans)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
