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
	"compress/flate"
	"io"
	"log"
	"os"
	"sync"

	"github.com/opennota/zip"
)

type book struct {
	Archive          string
	Filename         string
	Offset           int64
	CompressedSize   int64 `db:"compressed_size"`
	UncompressedSize int64 `db:"uncompressed_size"`
	fb2desc
	CRC32 uint32
	ID    uint32
}

type readCloser struct {
	io.Reader
	io.Closer
}

func (b *book) Open() (io.ReadCloser, error) {
	f, err := os.Open(b.Archive)
	if err != nil {
		return nil, err
	}

	sr := io.NewSectionReader(f, b.Offset, b.CompressedSize)

	return readCloser{sr, f}, nil
}

func (b *book) OpenDeflate() (io.ReadCloser, error) {
	r, err := b.Open()
	if err != nil {
		return nil, err
	}

	fr := flate.NewReader(r)

	return readCloser{fr, r.(readCloser).Closer}, nil
}

func indexZIP(name string, results chan<- book) error {
	r, err := zip.OpenReader(name)
	if err != nil {
		return err
	}
	defer r.Close()

	jobs := make(chan *zip.File, *parallel)
	var wg sync.WaitGroup
	wg.Add(*parallel)
	for i := 0; i < *parallel; i++ {
		go func() {
			defer wg.Done()

			for f := range jobs {
				rc, err := f.Open()
				if err != nil {
					log.Printf("%s/%s: open: %v", name, f.Name, err)
					continue
				}

				desc, err := ParseDesc(rc)
				if err == ErrSkip {
					continue
				}
				if err != nil {
					log.Printf("%s/%s: FB2 description: %v", name, f.Name, err)
					continue
				}

				offset, _ := f.DataOffset()
				results <- book{
					Archive:          name,
					Filename:         f.Name,
					fb2desc:          *desc,
					Offset:           offset,
					CompressedSize:   int64(f.CompressedSize64),
					UncompressedSize: int64(f.UncompressedSize64),
					CRC32:            f.CRC32,
				}
			}
		}()
	}

	for _, f := range r.File {
		if f.Method == zip.Deflate {
			jobs <- f
		} else {
			log.Printf("%s/%s: unsupported compression method", name, f.Name)
		}
	}

	close(jobs)
	wg.Wait()

	return nil
}
