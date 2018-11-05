# Chip8
A [Chip8][1] emulator written in Go.

Written as a learning exercise to get familiar with emulation, opcodes and CPUs.

## Requirements
The emulator uses SDL for it's graphics, input and sounds. Specifically it uses
the [Go SDL2][2] wrapper. This has it's own set of requirements so follow the
instructions on the [README][3].

## Usage
```bash
$ go run *.go
```

## References
As this was a learning exercise I had to seek a lot of help from the interwebs:
* [https://medium.com/average-coder/exploring-emulation-in-go-chip-8-636f99683f2a][4]
* [http://www.multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter][5]
* [https://en.wikipedia.org/wiki/CHIP-8#Virtual_machine_description][6]

[1]: https://en.wikipedia.org/wiki/CHIP-8
[2]: https://github.com/veandco/go-sdl2
[3]: https://github.com/veandco/go-sdl2#requirements
[4]: https://medium.com/average-coder/exploring-emulation-in-go-chip-8-636f99683f2a
[5]: http://www.multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter
[6]: https://en.wikipedia.org/wiki/CHIP-8#Virtual_machine_description