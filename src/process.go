// process.go -- process files/dirs and compute hashes
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
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/opencoff/go-utils"
	"github.com/opencoff/go-walk"
	"hash"
	"runtime"
)

const _parallelism int = 2

var nWorkers = runtime.NumCPU() * _parallelism

func processRecursive(args []string, h func() hash.Hash, fd io.Writer, opt *walk.Options) {
	walkCh, errch := walk.Walk(args, walk.FILE, opt)

	processChan(walkCh, errch, h, fd)
}

// iterate over the names
func processNormal(args []string, h func() hash.Hash, fd io.Writer, followSymlinks bool) {
	ch := make(chan walk.Result, nWorkers)
	errch := make(chan error, 1)

	// iterate in the background and feed the workers
	go func() {
		for _, nm := range args {
			var fi os.FileInfo
			var err error

			fi, err = os.Lstat(nm)
			if err != nil {
				errch <- fmt.Errorf("lstat %s: %w", nm, err)
				continue
			}

			m := fi.Mode()

			// if we're following symlinks, update fi & m
			if (m&os.ModeSymlink) > 0 && followSymlinks {
				if fi, err = os.Stat(nm); err != nil {
					errch <- fmt.Errorf("stat %s: %w", nm, err)
					continue
				}

				m = fi.Mode()
			}

			switch {
			case m.IsDir():
				Warn("skipping dir %s..", nm)

			case m.IsRegular():
				ch <- walk.Result{Path: nm, Stat: fi}

			default:
				Warn("skipping non-file %s..", nm)
			}
		}

		close(ch)
	}()

	// now handle the workers to hash em files
	processChan(ch, errch, h, fd)

	// this terminates the error harvesting goroutine in processChan below
	// XXX This feels v ugly. Fix it.
	close(errch)
}

func processChan(wch chan walk.Result, errch chan error, h func() hash.Hash, fd io.Writer) {
	var errs []error
	var wg sync.WaitGroup

	wg.Add(nWorkers)
	for i := 0; i < nWorkers; i++ {
		go func() {
			worker(wch, errch, h, fd)
			wg.Done()
		}()
	}

	go func(ch chan error) {
		for err := range ch {
			errs = append(errs, err)
		}
	}(errch)

	wg.Wait()

	if len(errs) > 0 {
		var b strings.Builder

		for _, err := range errs {
			b.WriteString(fmt.Sprintf("%s\n", err))
		}
		Die("%s", b)
	}
}

func worker(ch chan walk.Result, errch chan error, h func() hash.Hash, out io.Writer) {
	for r := range ch {
		if err := hashFile(r, h, out); err != nil {
			errch <- err
		}
	}
}

func hashFile(r walk.Result, hgen func() hash.Hash, out io.Writer) error {
	fd, err := os.Open(r.Path)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := hgen()
	if h == nil {
		panic(fmt.Sprintf("nil hash!"))
	}
	sz, err := utils.MmapReader(fd, 0, 0, h)
	if err != nil {
		return err
	}

	sum := h.Sum(nil)[:]

	_, err = fmt.Fprintf(out, "%x %d '%s'\n", sum, sz, r.Path)
	return err
}
