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
	file      string
	size      int64
	expsum    string
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

	var wg sync.WaitGroup
	ch := make(chan datum, nWorkers)
	errch := make(chan error, 1)

	// start workers that verify the hashes
	wg.Add(nWorkers)
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

	// feed the rest of the input file hash-lines
	wg.Add(1)
	go func(ch chan datum) {
		num := 2
		for ; rd.Scan(); num++ {
			errPref := fmt.Sprintf("%s: %d", nm, num)
			fn, sz, csum, err := parseLine(rd.Text(), errPref)
			if err != nil {
				errch <- err
				continue
			}

			ch <- datum{
				file:      fn,
				size:      sz,
				expsum:    csum,
				errPrefix: errPref,
			}

		}
		close(ch)
		wg.Done()
	}(ch)

	var errs []string
	var ewg sync.WaitGroup

	// harvest errors
	ewg.Add(1)
	go func(errch chan error) {
		for err := range errch {
			errs = append(errs, fmt.Sprintf("%s", err))
		}
		ewg.Done()
	}(errch)

	// don't reorder these:
	//  - we want to first wait for the workers to complete their hash verification
	//  - then, we close the error harvestor
	//  - and finally wait for error harvestor to complete
	//
	//  We can't read errs[] until the error harvestor has finished!
	wg.Wait()
	close(errch)
	ewg.Wait()

	if len(errs) > 0 {
		Warn("%s", strings.Join(errs, "\n"))
	}

	// return the exit code
	return 1 & len(errs)
}

func parseLine(line string, ep string) (fn string, sz int64, csum string, err error) {
	// fields are separated by '|'
	// field-1: hash
	// field-2: file size
	// field-3: quoted file name
	subs := strings.Split(line, "|")
	if len(subs) < 3 {
		err = fmt.Errorf("%s: malformed line; not enough fields", ep)
		return
	}

	wantHash := subs[0]
	sz, err = strconv.ParseInt(subs[1], 10, 64)
	if err != nil {
		err = fmt.Errorf("%s: malformed line; size %w", ep, err)
		return
	}

	fn, err = strconv.Unquote(subs[2])
	if err != nil {
		err = fmt.Errorf("%s: malformed line; filename %w", ep, err)
		return
	}

	// now we verify the file
	fi, err := os.Stat(fn)
	if err != nil {
		err = fmt.Errorf("%s: %w", ep, err)
		return
	}

	if !fi.Mode().IsRegular() {
		err = fmt.Errorf("%s: '%s' not a file", ep, fn)
		return
	}

	if fi.Size() != sz {
		err = fmt.Errorf("%s: '%s' size mismatch: exp %d, saw %d",
			ep, fn, sz, fi.Size())
		return
	}

	return fn, fi.Size(), wantHash, nil
}

func verifyFile(d datum, hgen func() hash.Hash) error {
	// finally we can hash and compare
	sum, sz, err := hashFile(d.file, hgen)
	if err != nil {
		return fmt.Errorf("%s: can't hash: %w", d.errPrefix, err)
	}

	// Account for hashFile() hashing fewer bytes
	if d.size != sz {
		return fmt.Errorf("%s: '%s' hash size mismatch: exp %d, saw %d",
			d.errPrefix, d.file, d.size, sz)
	}

	csum := fmt.Sprintf("%x", sum)
	if subtle.ConstantTimeCompare([]byte(csum), []byte(d.expsum)) != 1 {
		return fmt.Errorf("%s: file modified '%s'", d.errPrefix, d.file)
	}

	return nil
}
