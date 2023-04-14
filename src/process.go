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
	"strconv"
	"strings"
	"sync"

	"github.com/opencoff/go-walk"
	"hash"
	"runtime"
)

const _parallelism int = 2

var nWorkers = runtime.NumCPU() * _parallelism

func processRecursive(args []string, h func() hash.Hash, fd io.Writer, opt *walk.Options) {
	// both these chans will be closed by walk.Walk().
	// We can't just hand it off for further writes to processChan();
	// we need to create a smol goroutine to mirror to a different chan
	// that can be closed by processChan()


	// XXX Bug: symlinks that map to dirs are somehow returned back from Walk despite
	//	    followLinks being false.
	//          

	walkCh, errch := walk.Walk(args, walk.FILE, opt)

	out := make(chan walk.Result, nWorkers)
	err2 := make(chan error, 1)

	go func(in chan walk.Result, out chan walk.Result, errch chan error) {
		for w := range in {
			nm := w.Path
			fi := w.Stat
			m := fi.Mode()

			// if we're following symlinks, update fi & m
			if (m&os.ModeSymlink) > 0 {
				if !opt.FollowSymlinks {
					errch <- fmt.Errorf("skipping symlink %s", nm)
					continue
				}

				if fi, err = os.Stat(nm); err != nil {
					errch <- fmt.Errorf("stat %s: %w", nm, err)
					continue
				}

				m = fi.Mode()
			}

			switch {
			case m.IsDir():
				// XXX Add code to readlin() and create correct path to descend.
				panic("fix me")

			case m.IsRegular():
				out <- walk.Result{Path: nm, Stat: fi}

			default:
				errch <- fmt.Errorf("skipping non-file %s..", nm)
			}
		}
		close(out)

	}(walkCh, out, err2)

	go func() {
		for e := range errch {
			err2 <- e
		}
	}()

	// give our err2 chan for the workers
	processChan(walkCh, err2, h, fd)
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
			if (m&os.ModeSymlink) > 0 {
				if !followSymlinks {
					errch <- fmt.Errorf("skipping symlink %s", nm)
					continue
				}

				if fi, err = os.Stat(nm); err != nil {
					errch <- fmt.Errorf("stat %s: %w", nm, err)
					continue
				}

				m = fi.Mode()
			}

			switch {
			case m.IsDir():
				errch <- fmt.Errorf("skipping dir %s..", nm)

			case m.IsRegular():
				ch <- walk.Result{Path: nm, Stat: fi}

			default:
				errch <- fmt.Errorf("skipping non-file %s..", nm)
			}
		}

		close(ch)
	}()
}

	// now handle the workers to hash em files; processChan()
	// closes the error chan.
	processChan(ch, errch, h, fd)
}

func processChan(wch chan walk.Result, errch chan error, h func() hash.Hash, fd io.Writer) {
	var errs []error
	var wrkWait, errWait sync.WaitGroup

	wrkWait.Add(nWorkers)
	for i := 0; i < nWorkers; i++ {
		go func() {
			worker(wch, errch, h, fd)
			wrkWait.Done()
		}()
	}


	errWait.Add(1)
	go func(ch chan error) {
		for err := range ch {
			errs = append(errs, err)
		}
		errWait.Done()
	}(errch)

	wrkWait.Wait()

	close(errch)
	errWait.Wait()

	if len(errs) > 0 {
		var b strings.Builder

		for _, err := range errs {
			b.WriteString(fmt.Sprintf("%s\n", err))
		}
		Die("%s", b.String())
	}
}

func worker(ch chan walk.Result, errch chan error, h func() hash.Hash, out io.Writer) {
	for r := range ch {
		sum, sz, err := hashFile(r.Path, h)
		if err != nil {
			errch <- err
		} else {
			fn := strconv.Quote(r.Path)
			fmt.Fprintf(out, "%x|%d|%s\n", sum, sz, fn)
		}
	}
}