module github.com/opencoff/go-hash

go 1.20

require (
	github.com/opencoff/go-utils v0.4.1
	github.com/opencoff/go-walk v0.2.0
	github.com/opencoff/pflag v1.0.6-sh1
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/crypto v0.8.0
)

replace (
	github.com/opencoff/go-utils v0.4.1 => ../go-utils
	)
require (
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/term v0.7.0 // indirect
)
