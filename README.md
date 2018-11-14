# Chip8
A [Chip8][1] emulator written in Go.

Written as a learning exercise to get familiar with emulation, opcodes and CPUs.

## Building From Source
### Pre-requisites
The emulator uses the following packages which have requirements of their own
before we can build with them. Follow the instructions for each:

* [Pixel][2]
* [Oto][3]

Clone this repo and build the binary:
```bash
$ make
```

## Installation
```bash
$ go get -u github.com/danmrichards/chip8/cmd/chip8/...
```

## Usage
```bash
Usage of chip8:
  -debug
    	Run the emulator in debug mode
  -rom string
    	Path to the ROM file to load
```

## Controls
The Chip8 has a 16 key hex keyboard. For the purposes of this emulator it has
been implemented like so:
```
+---+---+---+---+
| 1 | 2 | 3 | 4 |
+---+---+---+---+
| Q | W | E | R |
+---+---+---+---+
| A | S | D | F |
+---+---+---+---+
| Z | X | C | V |
+---+---+---+---+
```
> Note: Which of these keys are actually used will differ from ROM to ROM.

## References
As this was a learning exercise I had to seek a lot of help from the interwebs:
* [https://medium.com/average-coder/exploring-emulation-in-go-chip-8-636f99683f2a][3]
* [http://www.multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter][4]
* [https://en.wikipedia.org/wiki/CHIP-8#Virtual_machine_description][5]

[1]: https://en.wikipedia.org/wiki/CHIP-8
[2]: https://github.com/faiface/pixel#requirements
[3]: https://github.com/hajimehoshi/oto#prerequisite
[4]: https://medium.com/average-coder/exploring-emulation-in-go-chip-8-636f99683f2a
[5]: http://www.multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter
[6]: https://en.wikipedia.org/wiki/CHIP-8#Virtual_machine_description