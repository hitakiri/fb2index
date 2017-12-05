// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	dataSource = flag.String("db", "file::memory:?cache=shared", "SQLite database")
	addr       = flag.String("http", "127.0.0.1:8080", "HTTP service address")
	recursive  = flag.Bool("r", false, "Recursively search for .zip files")
	parallel   = flag.Int("j", runtime.NumCPU(), "Number of parallel jobs")
	languages  = flag.String("l", "", "Comma-separated languages (default: all)")

	booksPerPage     = flag.Int("bpp", 50, "Books per page")
	authorsPerPage   = flag.Int("app", 50, "Authors per page")
	sequencesPerPage = flag.Int("spp", 50, "Sequences per page")

	cssPath = flag.String("css", "", "Use CSS file")

	allowedLanguages []string

	indexed int
)

func isRegular(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeType == 0
}

func walker(index func(string)) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, _ error) error {
		if !isRegular(fi) {
			return nil
		}

		if !strings.HasSuffix(path, ".zip") {
			return nil
		}

		index(path)

		return nil
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	if *languages != "" {
		allowedLanguages = strings.Split(*languages, ",")
	}

	initDB()
	ch, done := startInsertWorker()

	index := func(name string) {
		start := time.Now()
		err := indexZIP(name, ch)
		if err != nil {
			log.Printf("%s: open: %v", name, err)
		} else {
			log.Printf("Indexed %s in %v\n", name, time.Since(start))
			indexed++
		}
	}

	start := time.Now()

	for _, path := range flag.Args() {
		if *recursive {
			fi, err := os.Stat(path)
			if err != nil {
				log.Printf("%s: stat: %v", path, err)
				continue
			}

			if fi.IsDir() {
				filepath.Walk(path, walker(index))
				continue
			}
		}

		if !strings.HasSuffix(path, ".zip") {
			log.Printf("%s: not a .zip file", path)
			continue
		}

		index(path)
	}

	close(ch)
	<-done

	log.Printf("Indexed %d file(s) in %v", indexed, time.Since(start))
	log.Printf("Server listening on %s", *addr)
	listenAndServe()
}
