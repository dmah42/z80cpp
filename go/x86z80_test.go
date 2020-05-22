package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLabelRe(t *testing.T) {
	cases := []struct {
		line string
		want []string
	}{
		{},
		{
			line: "",
		},
		{
			line: "        mov    [22583], 32",
		},
		{
			line: "foo:",
			want: []string{"foo:", "foo"},
		},
	}

	for _, tt := range cases {
		if got := labelRE.FindStringSubmatch(tt.line); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("got %#v, want %#v", got, tt.want)
		}
	}
}

func TestOpRe(t *testing.T) {
	cases := []struct {
		line string
		want []string
	}{
		{},
		{
			line: "",
		},
		{
			line: "foo:",
		},
		{
			line: "mov     [22583], 32",
			want: []string{
				"mov     [22583], 32",
				"mov",
				"[22583], 32",
			},
		},
		{
			line: "mov     byte ptr [rcx + 22530], 32",
			want: []string{
				"mov     byte ptr [rcx + 22530], 32",
				"mov",
				"byte ptr [rcx + 22530], 32",
			},
		},
		{
			line: "jmp     .LBB0_1",
			want: []string{
				"jmp     .LBB0_1",
				"jmp",
				".LBB0_1",
			},
		},
		{
			line: "mov     dl, -2",
			want: []string{
				"mov     dl, -2",
				"mov",
				"dl, -2",
			},
		},
	}

	for _, tt := range cases {
		if got := opRE.FindStringSubmatch(tt.line); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("got %#v, want %#v", got, tt.want)
		}
	}
}

func TestConvertArg(t *testing.T) {
	cases := []struct {
		arg     string
		want    string
		wantMem bool
	}{
		{
			// register
			arg:  "eax",
			want: "a",
		},
		{
			// immediate
			arg:  "44234",
			want: "#44234",
		},
		{
			// immediate memory
			arg:     "byte ptr [22543]",
			want:    "#0x580f",
			wantMem: true,
		},
		{
			// register memory
			arg:     "byte ptr [ecx]",
			want:    "b",
			wantMem: true,
		},
	}

	for _, tt := range cases {
		got, gotMem := convertArg(tt.arg)
		if got != tt.want || gotMem != tt.wantMem {
			t.Errorf("%q: got %q %t, want %q %t", tt.arg, got, gotMem, tt.want, tt.wantMem)
		}
	}
}

func TestStripComments(t *testing.T) {
	cases := []struct {
		line, want string
	}{
		{},
		{
			line: "",
			want: "",
		},
		{
			line: "        mov   [22583], 32",
			want: "mov   [22583], 32",
		},
		{
			line: "        mov   [22583], 32  # foo",
			want: "mov   [22583], 32",
		},
	}

	for _, tt := range cases {
		if got := stripComments(tt.line); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}

func TestFormatLabel(t *testing.T) {
	cases := []struct {
		label, want string
	}{
		{
			want: ":\n",
		},
		{
			label: "foo",
			want:  "foo:\n",
		},
	}

	for _, tt := range cases {
		if got := formatLabel(tt.label); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}

func TestFormatOp(t *testing.T) {
	cases := []struct {
		operator string
		operands []string
		want     string
		wantErr  error
	}{
		{
			operator: "foo",
			wantErr:  fmt.Errorf("operator not found in mapping: %q", "foo"),
		},
		{
			operator: "mov",
			operands: []string{
				"[22583]", "32",
			},
			want: "        ld #22583, #32\n",
		},
		{
			operator: "jmp",
			operands: []string{".LBB01"},
			want:     "        jp .LBB01\n",
		},
	}

	for _, tt := range cases {
		got, err := formatOp(tt.operator, tt.operands)
		if !reflect.DeepEqual(err, tt.wantErr) {
			t.Errorf("got err %q, want err %q", err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}
