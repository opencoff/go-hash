// verify.go -- verify a list of hashes against entries in the filesys
//
// (c) 2023 Sudhi Herle <sudhi@herle.net>
//
// Licensing Terms: GPLv2
//
// If you need a commercial license for this work, please contact
// the author.
//
// This software does not come with any express or implied
// warranty; it is provided "as is". No claim  is made to its
// suitability for any purpose.

package main

import (
	"bufio"
	"fmt"
	"hash"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"crypto/subtle"
)

type datum struct {
	line      string
	errPrefix string
}

func doVerify(nm string) int {
	var fd io.ReadCloser = os.Stdin
	if nm != "-" && len(nm) > 0 {
		fx, err := os.Open(nm)
		if err != nil {
			Die("can't open '%s': %s", err)
		}
		fd = fx
	}

	defer fd.Close()

	rd := bufio.NewScanner(fd)
	if ok := rd.Scan(); !ok {
		Die("%s: possibly corrupt; can't read first line", nm)
	}

	subs := strings.Split(rd.Text(), " ")
	if len(subs) < 3 {
		Die("%s: possibly corrupt; not enough fields in header", nm)
	}

	magic := subs[0]
	if magic != MAGIC {
		Die("%s: Not a ghash file", nm)
	}

	halgo := subs[1]
	hgen, ok := Hashes[halgo]
	if !ok {
		Die("%s: unsupported hash algo '%s'", nm, halgo)
	}

	// start workers
	var wg sync.WaitGroup

	wg.Add(nWorkers)

	ch := make(chan datum, nWorkers)
	errch := make(chan error, 1)

	for i := 0; i < nWorkers; i++ {
		go func(ch chan datum, errch chan error) {
			for d := range ch {
				if err := verifyFile(d, hgen); err != nil {
					errch <- err
				}
			}
			wg.Done()
		}(ch, errch)
	}

	// feed the rest of the lines
	go func(ch chan datum) {
		num := 2
		for ; rd.Scan(); num++ {
			ch <- datum{line: rd.Text(), errPrefix: fmt.Sprintf("%s: %d", nm, num)}
		}
		close(ch)
	}(ch)

	// harvest errors
	var errs []string
	go func(errch chan error) {
		for err := range errch {
			errs = append(errs, fmt.Sprintf("%s", err))
		}
	}(errch)

	wg.Wait()
	close(errch)

	if len(errs) > 0 {
		Warn("%s", strings.Join(errs, "\n"))
	}

	// return the exit code
	return 1 & len(errs)
}

func verifyFile(d datum, hgen func() hash.Hash) error {
	line := d.line

	// fields are separated by '|'
	// field-1: hash
	// field-2: file size
	// field-3: quoted file name
	subs := strings.Split(line, "|")
	if len(subs) < 3 {
		return fmt.Errorf("%s: malformed line; not enough fields", d.errPrefix)
	}

	wantHash := subs[0]
	sz, err := strconv.ParseInt(subs[1], 10, 64)
	if err != nil {
		return fmt.Errorf("%s: malformed line; size %s", d.errPrefix, err)
	}

	fn, err := strconv.Unquote(subs[2])
	if err != nil {
		return fmt.Errorf("%s: malformed line; filename %s", d.errPrefix, err)
	}

	// now we verify the file
	fi, err := os.Stat(fn)
	if err != nil {
		return fmt.Errorf("%s: %s", d.errPrefix, err)
	}

	if !fi.Mode().IsRegular() {
		return fmt.Errorf("%s: '%s' not a file", d.errPrefix, fn)
	}

	if fi.Size() != sz {
		return fmt.Errorf("%s: '%s' size mismatch: exp %d, saw %d",
			d.errPrefix, fn, sz, fi.Size())
	}

	// finally we can hash and compare
	sum, sz, err := hashFile(fn, hgen)
	if err != nil {
		return fmt.Errorf("%s: can't hash: %s", d.errPrefix, err)
	}

	// Account for hashFile() hashing fewer bytes
	if fi.Size() != sz {
		return fmt.Errorf("%s: '%s' hash size mismatch: exp %d, saw %d",
			d.errPrefix, fn, fi.Size(), sz)
	}

	haveHash := fmt.Sprintf("%x", sum)
	if subtle.ConstantTimeCompare([]byte(haveHash), []byte(wantHash)) != 1 {
		return fmt.Errorf("%s: file modified '%s'", d.errPrefix, fn)
	}

	return nil
}
