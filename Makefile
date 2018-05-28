# clang++ -S -mllvm --x86-asm-syntax=intel test.cc 

CXX=clang++-5.0
CXXFLAGS=-O3 -Wall -Werror -std=c++1z
AS=sdasz80
ASFLAGS=-xlos -g
LD=sdcc
LDFLAGS=-mz80 --no-std-crt0 --nostdlib --code-loc 0x8032 --data-loc 0x8200 -Wl -b_HEADER=0x8000

app.tap: app.bin
	appmake +zx --binfile ./app.bin --org 32768

app.bin: game.rel crt0.rel
	$(LD) $(LDFLAGS) $^
	makezxbin <game.ihx >app.bin

%.rel: %.z80
	$(AS) $(ASFLAGS) $(basename $*).rel $*.z80

crt0.rel: crt0.s
	$(AS) $(ASFLAGS) $@ $<

game.z80: game.s x86z80
	./x86z80 -in $< -out $@

%.s: %.cc
	$(CXX) $(CXXFLAGS) -S -mllvm --x86-asm-syntax=intel $< -o $@

x86z80: ./go/x86z80.go
	go build -o $@ ./go

.PHONY: clean run test

run: app.tap
	fuse-gtk --tap $< \
	  --interface1 --interface2 \
	  --kempston --kempston-mouse \
	  --machine 48 --graphics-filter hq2x \
	  # --debugger-command "br 0x8000"

test: x86z80 ./go/x86z80_test.go
	go test ./go
	./x86z80 -in test.asm

clean:
	-@rm app.tap app.bin \
	  *.rel *.z80 *.lst *.sym .\
	  game.s game.map game.noi game.lk game.ihx \
	  /x86z80

