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

	// TODO: translate registers.
	trMap = map[string]func(...string) []string{
		"mov": func(args ...string) []string {
			var ops []string
			if _, err := strconv.Atoi(args[0]); err == nil {
				args[0] = "#" + args[0]
			} else if strings.HasPrefix(args[0], "byte ptr") {
				addr := dectohex(args[0][10 : len(args[0])-1])
				ops = append(ops, fmt.Sprintf("ld hl, #0x%s", addr))
				args[0] = "(hl)"
			}
			if _, err := strconv.Atoi(args[1]); err == nil {
				args[1] = "#" + args[1]
			} else if strings.HasPrefix(args[1], "byte ptr") {
				addr := dectohex(args[1][10 : len(args[0])-1])
				ops = append(ops, fmt.Sprintf("ld hl, #0x%s", addr))
				args[1] = "(hl)"
			}
			ops = append(ops, []string{
				// Z80: LD loads the second arg into the first.
				// Either arg can be a register, an 8bit integer, (HL) (the contents of register HL),
				// or (IX + d) or (IY + d) where d is a two's complement integer offset.
				// fmt.Sprintf("ld a, %s", args[1]),
				fmt.Sprintf("ld %s, %s", args[0], args[1]),
			}...)
			return ops
		},
		"and": func(args ...string) []string {
			return []string{
				// Z80: arg can be a register, a constant (8-bit), (HL), (IX+d), or (IY+d), d being 8-bit.
				fmt.Sprintf("ld a, %s", args[0]),
				fmt.Sprintf("and %s", args[1]),
				fmt.Sprintf("ld %s, a", args[0]),
			}
		},
		"shl": func(args ...string) []string {
			return []string{
				// Z80: arg can be the usual set.
				fmt.Sprintf("ld a, %s", args[0]),
				fmt.Sprintf("sla a"),
				fmt.Sprintf("ld %s, a", args[0]),
			}
		},
		"shr": func(args ...string) []string {
			return []string{
				// Z80: arg can be the usual set.
				fmt.Sprintf("ld a, %s", args[0]),
				fmt.Sprintf("sra a"),
				fmt.Sprintf("ld %s, a", args[0]),
			}
		},
		"jmp": func(args ...string) []string {
			return []string{
				// Z80: JP takes a 16-bit integer address or (HL), (IX), or (IY), unconditionally.
				fmt.Sprintf("jp %s", args[0]),
			}
		},
	}
)

func dectohex(dec string) string {
	addr, err := strconv.Atoi(dec)
	if err != nil {
		return fmt.Sprintf("<Unable to convert %q>", dec)
	}
	return fmt.Sprintf("%x", addr)
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
