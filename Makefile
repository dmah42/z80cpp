# clang++ -S -mllvm --x86-asm-syntax=intel test.cc 

CXX=clang++-5.0
CXXFLAGS=-O3 -Wall -Werror -std=c++1z
AS=sdasz80
ASFLAGS=-xlos -g
LD=sdcc
LDFLAGS=-mz80 --no-std-crt0 --nostdlib --code-loc 0x8032 --data-loc 0x8200 -Wl -b_HEADER=0x8000

OUT=out

all: $(OUT)/app.tap

%.tap: %.bin
	appmake +zx --binfile $< --org 32768

$(OUT)/app.bin: $(OUT)/game.rel $(OUT)/crt0.rel
	cd $(@D); \
	  $(LD) $(LDFLAGS) $(^F); \
	  makezxbin <game.ihx >$(@F)

%.rel: %.z80
	$(AS) $(ASFLAGS) $@ $<

$(OUT)/crt0.rel: crt0.s
	@mkdir -p $(@D)
	$(AS) $(ASFLAGS) $@ $<

$(OUT)/game.z80: game.s $(OUT)/x86z80
	@mkdir -p $(@D)
	$(OUT)/x86z80 -in $< -out $@

%.s: %.cc
	$(CXX) $(CXXFLAGS) -S -mllvm --x86-asm-syntax=intel $< -o $@

$(OUT)/x86z80: ./go/x86z80.go
	@mkdir -p $(@D)
	go build -o $@ ./go

run: $(OUT)/app.tap
	fuse-gtk --tap $< \
	  --interface1 --interface2 \
	  --kempston --kempston-mouse \
	  --machine 48 --graphics-filter hq2x \
	  # --debugger-command "br 0x8000"

test: $(OUT)/x86z80 ./go/x86z80_test.go
	go test ./go
	$< -in test.asm -out $(OUT)/test.z80

clean:
	-@rm -r $(OUT)

.PHONY: clean run test all

