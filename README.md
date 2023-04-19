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
Usage:

	ghash [options] file|dir [file|dir ..]

	Options:
	  -h, --help            Show help and exit
	  -V, --version         Show version info and exit
	  -r, --recurse	        Recursively traverse directories
	  -x, --one-filesystem  Don't cross file system boundaries
	  -L, --follow-symlinks Follow symbolic links
	  -H, --hash=H		Use hash algorithm 'H' [sha256]
	  --list-hashes		List supported hash algorithms
	  -v, --verify-from=F   Verify the hashes in file 'F' [stdin]
	  -o, --output=O        Write output hashes to file 'O' [stdout]

### Hashing individual files

    ghash file1 file2 file3

### Hashing recursively multiple files and dirs

    ghash -r dir1 dir2 file3 file4

### Verifying previously generated hashes

    # first generate and save the hashes
    ghash -L -x -H blake3 -o /tmp/etc.ghash -r /etc

    # now verify the saved hashes in /tmp/etc.ghash
    ghash -v /tmp/etc.ghash

## Licensing Terms
The tool and code is licensed under the terms of the
GNU Public License v2.0 (strictly v2.0). If you need a commercial
license or a different license, please get in touch with me.

See the file ``LICENSE.md`` for the full terms of the license.

## Author
Sudhi Herle <sw@herle.net>

