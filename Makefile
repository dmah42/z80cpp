# clang++ -S -mllvm --x86-asm-syntax=intel test.cc 

CXX=clang++-3.5
CXXFLAGS=-O3 -Wall -Werror -std=c++1z

game.z80: game.s x86z80
	./x86z80 -in game.s

x86z80: ./go/x86z80.go
	go build -o x86z80 ./go

game.s: game.cc
	$(CXX) $(CXXFLAGS) -S -mllvm --x86-asm-syntax=intel game.cc

test: x86z80 ./go/x86z80_test.go
	go test ./go
	./x86z80 -in test.asm

clean:
	-rm game.z80 game.s test.z80 ./x86z80

