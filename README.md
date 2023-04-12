# README for go-hash

## What is this?
Pure go tool for calculating and verifying hashes of files & dirs.

It exploits concurrency where possible and uses mmap(2) for
hashing/verifying files.

The hash output can be saved in a file for future verification.


## How do I build it?
You need a modern go toolchain (go 1.17+):


    git clone https://github.com/opencoff/go-hash
    cd go-hash
    make

The binary will be in `./bin/$HOSTOS-$ARCH/ghash`.
where `$HOSTOS` is the host OS where you are building (e.g., openbsd)
and `$ARCH` is the CPU architecture (e.g., amd64).

## How do I use it?


## Licensing Terms
The tool and code is licensed under the terms of the
GNU Public License v2.0 (strictly v2.0). If you need a commercial
license or a different license, please get in touch with me.

See the file ``LICENSE.md`` for the full terms of the license.

## Author
Sudhi Herle <sw@herle.net>

