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
	"io"
	"os"
	/*
		"fmt"
		"sync"
		"strings"

		"github.com/opencoff/go-walk"
		"github.com/opencoff/go-utils"
		"hash"
		"runtime"
	*/)

func doVerify(nm string) {
	var fd io.ReadCloser = os.Stdin
	if nm != "-" && len(nm) > 0 {
		fx, err := os.Open(nm)
		if err != nil {
			Die("can't open '%s': %s", err)
		}
		fd = fx
	}

	defer fd.Close()

	rd := bufio.NewReader(fd)
	line, tooLong, err := rd.ReadLine()
	if toolong {
		Die("%s: possibly corrupt; first line too long", nm)
	}

	if err != nil {
		Die("%s: %s", nm, err)
	}

	line := string(line)
	want := len(MAGIC)
	if len(line) < want  || line[:want] != MAGIC {
		Die("%s: Not a ghash file", nm)
	}

	line = line[want:]
	subs := strings.Split(line, " ")
	if len(subs) < 2 {
		Die("%s: Corrupted ghash file", nm)
	}

	hgen, ok := Hashes[subs[0]]
	if !ok {
		Die("%s: unsupported hash algo %s", subs[0])
	}


	// start workers
	var wg sync.WaitGroup

	wg.Add(nWorkers)

	type datum struct {
		line string
		errPrefix string
	}
	ch := make(chan datum, nWorkers)
	errch := make(chan error, 1)

	for i := 0; i < nWorkers; i++ {
		go func() {
			worker(ch, errch)
			wg.Done()
		}
	}

	// feed the rest of the lines
	go func() {
		num := 2
		for ;; num++{
			line, tooLong, err := rd.ReadLine()
			if toolong {
				errch <- fmt.Errorf("%s: %d: line too long; skipped.", nm, num)
				continue
			}

			if err != nil {
				errch <- fmt.Errorf("%s: %d: %s", nm, num, err)
				continue
			}
			ch <- datum{line: line, errPrefix: fmt.Sprintf("%s: %d", nm, num)}
		}
	}()

	// harvest errors
	var errs []string
	go func() {
		for err := range errch {
			errs = append(errs, fmt.Sprintf("%s", err))
		}
	}()

	wg.Wait()

	// this ends the goroutine above
	close(errch)

	if len(errs) > 0 {
		Warn("%s", strings.Join(errs, "\n"))
	}
}


func worker(ch chan datum, errch chan error) {
	for d := range ch {
		line := d.line
		i := strings.Index(line, " ")
		if i < 0 {
			errch <- fmt.Errorf("%s: malformed line; no hash", d.errPrefix)
			continue
		}
		hash := line[:i]
	}
}
