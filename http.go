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
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var rImageName = regexp.MustCompile(`^(\d+)_(\d+)\.(?:jpg|jpeg|png|gif)$`)

func logError(r *http.Request, err error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	log.Println(err, "--", host, r.Method, r.URL, r.Referer(), r.UserAgent())
}

func httpError(w http.ResponseWriter, r *http.Request, err error) {
	logError(r, err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func executeTemplate(w http.ResponseWriter, name string, data interface{}) error {
	tmpl := templates[name]
	if tmpl == nil {
		return fmt.Errorf("template %s not found", name)
	}

	w.Header().Add("Content-Type", "text/html")
	return tmpl.ExecuteTemplate(w, "base", data)
}

func intFormValue(r *http.Request, name string) int {
	value, err := strconv.Atoi(r.FormValue(name))
	if err != nil {
		return -1
	}

	return value
}

func intFormValueDefault(r *http.Request, name string, defaultValue int) int {
	vs := r.Form[name]
	if len(vs) == 0 {
		return defaultValue
	}

	value, err := strconv.Atoi(vs[0])
	if err != nil {
		return -1
	}

	return value
}

func ID(r *http.Request) uint32 {
	id := intFormValue(r, "id")
	if id < 0 {
		return 0
	}
	return uint32(id)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	bookHandler(w, r)
}

func bookDownload(w http.ResponseWriter, b *book) error {
	r, err := b.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	w.Header().Add("Content-Type", "application/fb2")
	w.Header().Add("Content-Encoding", "deflate")
	w.Header().Add("Content-Length", fmt.Sprint(b.CompressedSize))
	w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%d.fb2"`, b.ID))
	_, err = io.Copy(w, r)

	return err
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	if id := ID(r); id > 0 {
		switch r.FormValue("action") {
		case "read":
			b, err := BookByID(id)
			if err == ErrNoRows {
				http.NotFound(w, r)
				return
			}
			if err != nil {
				httpError(w, r, err)
				return
			}

			html, err := b.HTML()
			if err != nil {
				httpError(w, r, err)
				return
			}

			err = executeTemplate(w, "book_read", struct {
				Book *book
				HTML template.HTML
			}{
				b,
				template.HTML(html),
			})
			if err != nil {
				logError(r, err)
				return
			}
		case "download":
			b, err := BookByID(id)
			if err == ErrNoRows {
				http.NotFound(w, r)
				return
			}
			if err != nil {
				httpError(w, r, err)
				return
			}

			err = bookDownload(w, b)
			if err != nil {
				logError(r, err)
				return
			}
		default:
			b, ann, cover, err := BookByIDWithAnnotation(id)
			if err == ErrNoRows {
				http.NotFound(w, r)
				return
			}
			if err != nil {
				httpError(w, r, err)
				return
			}

			err = executeTemplate(w, "book", struct {
				Book     *book
				Ann      template.HTML
				Cover    string
				Language string
			}{
				b,
				template.HTML(ann),
				cover,
				iso639_1[b.Lang],
			})
			if err != nil {
				logError(r, err)
				return
			}
		}

		return
	}

	page := intFormValueDefault(r, "page", 1)
	if page <= 0 {
		http.NotFound(w, r)
		return
	}

	books, totalPages, err := BooksPerPage(page)
	if err != nil {
		httpError(w, r, err)
		return
	}

	err = executeTemplate(w, "book_index", struct {
		Books      []book
		PageNumber int
		TotalPages int
	}{
		books,
		page,
		totalPages,
	})
	if err != nil {
		logError(r, err)
		return
	}
}

func authorHandler(w http.ResponseWriter, r *http.Request) {
	if id := ID(r); id > 0 {
		books, translations, au, err := BooksByAuthor(id)
		if err == ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			httpError(w, r, err)
			return
		}

		err = executeTemplate(w, "author", struct {
			Author       *author
			Books        []book
			Translations []book
		}{
			au,
			books,
			translations,
		})
		if err != nil {
			logError(r, err)
			return
		}

		return
	}

	page := intFormValueDefault(r, "page", 1)
	if page <= 0 {
		http.NotFound(w, r)
		return
	}

	authors, totalPages, err := AuthorsPerPage(page)
	if err != nil {
		httpError(w, r, err)
		return
	}

	err = executeTemplate(w, "author_index", struct {
		Authors    []author
		PageNumber int
		TotalPages int
	}{
		authors,
		page,
		totalPages,
	})
	if err != nil {
		logError(r, err)
		return
	}
}

func sequenceHandler(w http.ResponseWriter, r *http.Request) {
	if id := ID(r); id > 0 {
		books, seq, err := BooksBySequence(id)
		if err == ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			httpError(w, r, err)
			return
		}

		err = executeTemplate(w, "sequence", struct {
			Sequence *sequence
			Books    []book
		}{
			seq,
			books,
		})
		if err != nil {
			logError(r, err)
			return
		}

		return
	}

	page := intFormValueDefault(r, "page", 1)
	if page <= 0 {
		http.NotFound(w, r)
		return
	}

	sequences, totalPages, err := SequencesPerPage(page)
	if err != nil {
		httpError(w, r, err)
		return
	}

	err = executeTemplate(w, "sequence_index", struct {
		Sequences  []sequence
		PageNumber int
		TotalPages int
	}{
		sequences,
		page,
		totalPages,
	})
	if err != nil {
		logError(r, err)
		return
	}
}

func genreHandler(w http.ResponseWriter, r *http.Request) {
	if id := ID(r); id > 0 {
		books, g, err := BooksByGenre(id)
		if err == ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			httpError(w, r, err)
			return
		}

		err = executeTemplate(w, "genre", struct {
			Genre *genre
			Books []book
		}{
			g,
			books,
		})
		if err != nil {
			logError(r, err)
			return
		}

		return
	}

	genres, err := Genres()
	if err != nil {
		httpError(w, r, err)
		return
	}

	err = executeTemplate(w, "genre_index", genres)
	if err != nil {
		logError(r, err)
		return
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		err := executeTemplate(w, "search", nil)
		if err != nil {
			logError(r, err)
			return
		}
	case "POST":
		query := r.FormValue("query")
		authors, sequences, books, err := Search(query)
		if err != nil {
			httpError(w, r, err)
			return
		}

		err = executeTemplate(w, "search", struct {
			Authors     []author
			Sequences   []sequence
			Books       []book
			SearchQuery string
		}{
			authors,
			sequences,
			books,
			query,
		})
		if err != nil {
			logError(r, err)
			return
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func contentTypeByExt(name string) string {
	switch strings.ToLower(path.Ext(name)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	}
	return ""
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[3:]

	if name == "no-cover.png" {
		w.Header().Add("Content-Type", "image/png")
		w.Write(nocoverpng)
		return
	}

	if !rImageName.MatchString(name) {
		http.NotFound(w, r)
		return
	}

	contentType := contentTypeByExt(name)
	if contentType == "" {
		http.NotFound(w, r)
		return
	}

	data := imageCache.Get(name)
	if data == nil {
		submatches := rImageName.FindStringSubmatch(name)

		id, err := strconv.Atoi(submatches[1])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		sum, err := strconv.Atoi(submatches[2])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		b, err := BookByID(uint32(id))
		if err == ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			httpError(w, r, err)
			return
		}

		data, err = b.Image(uint32(sum))
		if data == nil {
			if err != nil {
				httpError(w, r, err)
				return
			}
			http.NotFound(w, r)
			return
		}

		imageCache.Put(name, data)
	}

	w.Header().Add("Content-Type", contentType)
	_, err := w.Write(data)
	if err != nil {
		logError(r, err)
		return
	}
}

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}

func listenAndServe() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/b", bookHandler)
	http.HandleFunc("/g", genreHandler)
	http.HandleFunc("/a", authorHandler)
	http.HandleFunc("/s", sequenceHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/i/", imageHandler)
	http.HandleFunc("/robots.txt", robotsHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var nocoverpng = []byte(
	"\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48\x44\x52" +
		"\x00\x00\x00\x40\x00\x00\x00\x40\x08\x00\x00\x00\x00\x8f\x02\x2e" +
		"\x02\x00\x00\x01\x5d\x49\x44\x41\x54\x58\xc3\xed\xd6\xfd\x6e\x82" +
		"\x30\x10\x00\x70\xdf\xff\x4d\x28\x5f\x03\xe7\x18\x66\x24\x48\x14" +
		"\x34\x01\x04\x07\x08\x14\xdf\x65\x19\xd2\xab\x19\x6e\x39\x6e\xc9" +
		"\x3e\x12\xee\xcf\x4b\xfb\x13\xca\xf5\xce\xc5\xe5\x9b\xb1\x98\x81" +
		"\x19\xf8\xe3\x00\x4f\xb7\xae\x6d\x68\x8a\x66\x2e\xbd\x30\xe7\x53" +
		"\x81\x93\xa7\xb1\xdb\x30\x36\xd5\x14\xa0\x5c\xb3\x71\x04\x2d\x1a" +
		"\x88\x35\x76\x2f\x56\x0d\x12\x88\x14\x76\x3f\xec\x16\x05\x1c\xc5" +
		"\x7e\xc3\xcf\xaa\x26\x8f\x0c\x29\xf8\x18\x80\x9b\xe2\x89\xeb\x6b" +
		"\xa2\x71\xa5\x50\x21\x80\x83\x58\x5c\x88\x4c\x23\x9f\x61\x8b\x00" +
		"\x9c\x11\x70\x09\xe5\x39\x22\x00\xf1\x06\x6c\x07\xa9\x0c\x00\x13" +
		"\x01\xa8\x62\xf1\x0b\xa4\x1a\x00\x54\x04\x60\x89\xc5\x0e\xa4\x6a" +
		"\x00\x34\x04\xe0\x8d\xbf\x59\x0e\x80\x85\x00\xaa\xe1\x1d\xd4\x12" +
		"\x52\x3e\x00\x1e\xaa\x90\xfa\x63\x7c\x48\x21\x51\xc8\x3a\x48\x51" +
		"\xa5\xdc\x26\xbb\x5d\x22\xab\xf6\x0c\xa7\xc2\x96\x9c\xd0\x50\x2a" +
		"\x1b\xf6\x2b\x39\xa1\x23\x15\x37\x57\x21\x22\xb4\xb4\x52\x97\xbf" +
		"\xbf\x27\xf4\xc4\x56\x3e\xff\xd3\x2b\xa5\xa9\xc2\x2d\xb0\xe2\x8e" +
		"\xd4\x95\x57\xc3\x15\x38\x70\x5a\x5b\xef\x86\x5b\xd1\x52\xe7\x42" +
		"\xdb\xef\x7f\xec\xc8\x83\xe5\xdc\x03\x7b\xfa\x64\xaa\x3f\xa9\x5f" +
		"\x34\xc0\x83\xf7\xa8\xfe\xf5\x74\x8e\x1d\xdd\x5c\x9f\xe8\xc0\xe6" +
		"\x63\x7b\x9d\x08\x24\xe2\x1a\x65\x44\x00\x46\xc4\x33\x11\x80\x21" +
		"\xad\x13\x01\x98\x31\x06\x11\x80\xa9\xea\x12\x81\xe3\x17\xbd\x10" +
		"\xf7\x19\x83\xeb\xfe\xf0\xf7\x0a\x69\xfe\xaf\x3c\x03\x33\xf0\x73" +
		"\xc0\x1b\xaf\xc4\x73\x97\xc2\x2e\x69\xa1\x00\x00\x00\x00\x49\x45" +
		"\x4e\x44\xae\x42\x60\x82")
