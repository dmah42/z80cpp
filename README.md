# z80cpp
z80 code in c++.

all z80 compilers are C (or you can use assemblers). I want to use C++.

strategy: write some C++ that optimizes beautifully (thanks, https://gcc.godbolt.org!), compile to x86, convert to z80.

it's that or write a compiler backend for z80 for llvm/clang.

## Instructions
```
$ make
```

Will compile `game.cc` to x86 assembly (on x86 platforms), compile the `x86z80` tool, and run it against the assembly. This will generate `game.z80` which can (in the future) be assembled into an executable for ZX spectrum.

```
$ make test
```

This will run the `x86z80` translator against the `test.asm` file. It does not attempt to verify the output.
