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
	"strconv"
	"strings"
)

var (
	in  = flag.String("in", "test.asm", "the file to read")
	out = flag.String("out", "test.z80", "the file to write")

	labelRE     = regexp.MustCompile(`^(\.?\w+):`)
	opRE        = regexp.MustCompile(`^(\w+)\s+(.*)`)
	directiveRE = regexp.MustCompile(`^\s*(\.\w+)(\s+.+)?`)
	addressOfRE = regexp.MustCompile(`byte ptr \[(.+)\]`)

	// TODO: translate registers through a map.
	trMap = map[string]func(...string) []string{
		// mov moves the second argument into the first.
		"mov": func(args ...string) (ops []string) {
			arg, isMem := convertArg(args[0])
			if isMem {
				ops = append(ops, fmt.Sprintf("ld hl, %s", arg))
				args[0] = "(hl)"
			} else {
				args[0] = arg
			}
			arg, isMem = convertArg(args[1])
			if isMem {
				ops = append(ops, fmt.Sprintf("ld de, %s", arg))
				args[1] = "(de)"
			} else {
				args[1] = arg
			}
			return append(ops, fmt.Sprintf("ld %s, %s", args[0], args[1]))
		},
		// z80 and does a bitwise and and puts the result into the accumulator.
		// x86 and does a bitwise and between src and dest and puts the result into dest.
		"and": func(args ...string) (ops []string) {
			arg, isMem := convertArg(args[0])
			if isMem {
				ops = append(ops, fmt.Sprintf("ld hl, %s", arg))
				args[0] = "(hl)"
			}
			arg, isMem = convertArg(args[1])
			if isMem {
				ops = append(ops, fmt.Sprintf("ld de, %s", arg))
				args[1] = "(de)"
			}
			return append(ops, []string{
				fmt.Sprintf("ld a, %s", args[1]),
				fmt.Sprintf("and %s", args[0]),
				fmt.Sprintf("ld %s, a", args[1]),
			}...)
		},
		// "shl": func(args ...string) []string {
		// 	return []string{
		// 		// Z80: arg can be the usual set.
		// 		// TODO: parse and check memory vs register.
		// 		fmt.Sprintf("sla %s", args[0]),
		// 	}
		// },
		// "shr": func(args ...string) []string {
		// 	return []string{
		// 		// Z80: arg can be the usual set.
		// 		// TODO: parse and check memory vs register.
		// 		fmt.Sprintf("sra %s", args[0]),
		// 	}
		// },
		"jmp": func(args ...string) []string {
			return []string{
				// Z80: JP takes a 16-bit integer address or (HL), (IX), or (IY), unconditionally.
				fmt.Sprintf("jp %s", args[0]),
			}
		},
		"inc": func(args ...string) (ops []string) {
			// z80 inc increases operand (8- or 16-bit) by 1.
			arg, isMem := convertArg(args[0])
			if isMem {
				ops = append(ops, fmt.Sprintf("ld hl, %s", arg))
				args[0] = "(hl)"
			}
			return append(ops, fmt.Sprintf("inc %s", arg))
		},
	}
)

// Returns parsed arg and 'true' if the location is a memory address.
func convertArg(arg string) (string, bool) {
	if _, err := strconv.Atoi(arg); err == nil {
		return "#" + arg, false
	}

	// Address of.
	maybeAddressOf := addressOfRE.FindStringSubmatch(arg)
	if maybeAddressOf != nil {
		loc := maybeAddressOf[1]
		addr, err := strconv.Atoi(loc)
		if err != nil {
			return convertReg(loc), true
		}
		return fmt.Sprintf("#0x%x", addr), true
	}

	return convertReg(arg), false
}

func convertReg(reg string) string {
	if reg == "eax" {
		return "a"
	}
	if reg == "ecx" {
		return "b"
	}
	// TODO: something else
	return reg
}

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
			trans = fmt.Sprintf("\t%s\n", ops[0])
			for _, o := range ops[1:] {
				trans += fmt.Sprintf("\t%s\n", o)
			}
		}
		return trans, nil
	}
	return "", fmt.Errorf("operator not found in mapping: %q", operator)
}

func formatDirective(directive string, params []string) (string, error) {
	log.Println(directive, params)
	// TODO: map
	switch directive {
	case ".intel_syntax", ".file", ".p2align", ".cfi_startproc", ".cfi_endproc", ".type":
		return "", nil
	case ".text":
		return "\t.area\t_CODE\n", nil
	case ".globl":
		if len(params) != 1 {
			return "", fmt.Errorf("expected param length 1, got %d", len(params))
		}
		return fmt.Sprintf("\t%s\t%s\n", directive, params[0]), nil
	}

	return "", fmt.Errorf("directive not found: %q", directive)
}

func translate(line string) (string, error) {
	line = stripComments(line)
	if len(line) == 0 {
		return "", nil
	}
	maybeLabel := labelRE.FindStringSubmatch(line)
	if maybeLabel != nil {
		return formatLabel(maybeLabel[1]), nil
	}

	maybeOp := opRE.FindStringSubmatch(line)
	if maybeOp != nil {
		operator := strings.ToLower(maybeOp[1])
		operands := strings.Split(maybeOp[2], ",")
		for i, o := range operands {
			operands[i] = strings.TrimSpace(o)
		}
		return formatOp(operator, operands)
	}

	maybeDirective := directiveRE.FindStringSubmatch(line)
	if maybeDirective != nil {
		directive := strings.ToLower(maybeDirective[1])
		var params []string
		if len(maybeDirective) > 2 {
			params = strings.Split(maybeDirective[2], ",")
			for i, p := range params {
				params[i] = strings.TrimSpace(p)
			}
		}
		return formatDirective(directive, params)
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

	if *out == "" {
		*out = fmt.Sprintf("%s.z80", strings.TrimSuffix(absin, path.Ext(*in)))
	}
	log.Printf("opening %q for writing", *out)
	outfile, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	module := strings.TrimSuffix(path.Base(absin), path.Ext(*in))
	fmt.Fprint(outfile, fmt.Sprintf("\t.module %s\n", module))

	lineno := 1
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprint(outfile, fmt.Sprintf(";; %d: %s\n", lineno, line))
		trans, err := translate(line)
		if err != nil {
			log.Printf(err.Error())
			trans = fmt.Sprintf(";; ERROR: %s\n", err)
		}
		fmt.Fprint(outfile, trans)
		lineno += 1
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
