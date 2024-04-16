module github.com/opencoff/go-hash

go 1.21.1

require (
	github.com/opencoff/go-mmap v0.1.2
	github.com/opencoff/go-utils v0.9.3
	github.com/opencoff/go-walk v0.6.1
	github.com/opencoff/pflag v1.0.6-sh2
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/crypto v0.22.0
)

require (
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/term v0.19.0 // indirect
)

//replace github.com/opencoff/go-mmap => ../go-mmap
