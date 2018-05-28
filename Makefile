# clang++ -S -mllvm --x86-asm-syntax=intel test.cc 

CXX=clang++-5.0
CXXFLAGS=-O3 -Wall -Werror -std=c++1z
AS=sdasz80
ASFLAGS=-xlos -g
LD=sdcc
LDFLAGS=-mz80 --no-std-crt0 --nostdlib --code-loc 0x8032 --data-loc 0x8200 -Wl -b_HEADER=0x8000
OUT=out

$(OUT)/app.tap: $(OUT)/app.bin
	appmake +zx --binfile $(OUT)/app.bin --org 32768

$(OUT)/app.bin: game.rel crt0.rel
	$(LD) $(LDFLAGS) $^
	makezxbin <game.ihx >$@

%.rel: %.z80
	$(AS) $(ASFLAGS) $(basename $*).rel $*.z80

crt0.rel: crt0.s
	$(AS) $(ASFLAGS) $@ $<

game.z80: game.s $(OUT)/x86z80
	$(OUT)/x86z80 -in $< -out $@

%.s: %.cc
	$(CXX) $(CXXFLAGS) -S -mllvm --x86-asm-syntax=intel $< -o $@

$(OUT)/x86z80: ./go/x86z80.go
	go build -o $@ ./go

.PHONY: clean run test

run: $(OUT)/app.tap
	fuse-gtk --tap $< \
	  --interface1 --interface2 \
	  --kempston --kempston-mouse \
	  --machine 48 --graphics-filter hq2x \
	  # --debugger-command "br 0x8000"

test: $(OUT)/x86z80 ./go/x86z80_test.go
	go test ./go
	./x86z80 -in test.asm

clean:
	-@rm -r *.rel *.z80 *.lst *.sym \
	  game.s game.map game.noi game.lk game.ihx \
	  $(OUT)

