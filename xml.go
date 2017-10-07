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
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"hash/crc32"
	"html"
	"io"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/opennota/fb2index/cache"
	"github.com/rogpeppe/go-charset/charset"
	_ "github.com/rogpeppe/go-charset/data"
)

var (
	ErrNoTitle = errors.New("no title")
	ErrSkip    = errors.New("skip this book")
)

var imageCache = cache.New(time.Minute, 10*time.Second)

type genre struct {
	Name      string
	Desc      string
	Meta      string
	BookCount int
	ID        uint32
}

type author struct {
	FirstName  string `db:"first_name"`
	MiddleName string `db:"middle_name"`
	LastName   string `db:"last_name"`
	Nickname   string
	BookCount  int
	ID         uint32
}

type sequence struct {
	Name      string
	Number    int
	BookCount int
	ID        uint32
}

type fb2desc struct {
	Genres      []genre
	Authors     []author
	Translators []author
	Sequences   []sequence
	Title       string
	Lang        string
}

func skip(d *xml.Decoder, name xml.Name) error {
	lvl := 0

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name == name {
				lvl++
			}
		case xml.EndElement:
			if tok.Name == name {
				if lvl == 0 {
					return nil
				}
				lvl--
			}
		}
	}
}

func text(d *xml.Decoder, name xml.Name) string {
	var buf bytes.Buffer

	for {
		tok, err := d.Token()
		if err != nil {
			return ""
		}

		switch tok := tok.(type) {
		case xml.EndElement:
			return string(bytes.TrimSpace(buf.Bytes()))
		case xml.CharData:
			buf.Write(tok)
		default:
			return ""
		}
	}
}

func attr(e xml.StartElement, local string) string {
	for _, a := range e.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

func lower(b byte) bool {
	return 'a' <= b && b <= 'z'
}

func validLanguage(lang string) bool {
	return len(lang) == 2 && lower(lang[0]) && lower(lang[1])
}

func languageAllowed(lang string) bool {
	if len(allowedLanguages) == 0 {
		return validLanguage(lang)
	}
	for _, l := range allowedLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

func validGenre(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_') {
			return false
		}
	}
	return true
}

func ParseDesc(r io.Reader) (*fb2desc, error) {
	d := xml.NewDecoder(io.LimitReader(r, 16384))
	d.CharsetReader = charset.NewReader

	var a author
	var desc fb2desc
	inTitleInfo := false
loop:
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "FictionBook", "description":
				// do nothing
			case "title-info":
				inTitleInfo = true
			case "genre":
				g := normalizeGenre(text(d, tok.Name))
				if validGenre(g) {
					desc.Genres = append(desc.Genres, genre{
						Name: g,
					})
				}
			case "author", "translator":
				a = author{}
			case "first-name":
				a.FirstName = text(d, tok.Name)
			case "middle-name":
				a.MiddleName = text(d, tok.Name)
			case "last-name":
				a.LastName = text(d, tok.Name)
			case "nickname":
				a.Nickname = text(d, tok.Name)
			case "book-title":
				desc.Title = text(d, tok.Name)
			case "lang":
				desc.Lang = text(d, tok.Name)
				if !languageAllowed(desc.Lang) {
					return nil, ErrSkip
				}
			case "sequence":
				name := attr(tok, "name")
				if name != "" {
					number, _ := strconv.Atoi(attr(tok, "number"))
					desc.Sequences = append(desc.Sequences, sequence{
						Name:   name,
						Number: number,
					})
				}
			default:
				err := skip(d, tok.Name)
				if err != nil {
					break loop
				}
			}
		case xml.EndElement:
			switch tok.Name.Local {
			case "description":
				break loop
			case "title-info":
				inTitleInfo = false
			case "author":
				if inTitleInfo && (a != author{}) {
					desc.Authors = append(desc.Authors, a)
				}
			case "translator":
				if inTitleInfo && (a != author{}) {
					desc.Translators = append(desc.Translators, a)
				}
			}
		}
	}

	if desc.Title == "" {
		return nil, ErrNoTitle
	}

	return &desc, nil
}

func parseAnnotation(d *xml.Decoder) (string, error) {
	var buf bytes.Buffer

	for {
		tok, err := d.Token()
		if err != nil {
			return "", err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "p":
				buf.WriteString("<p>")
			case "emphasis":
				buf.WriteString("<em>")
			case "strong":
				buf.WriteString("<strong>")
			case "strikethrough":
				buf.WriteString("<strike>")
			case "empty-line":
				buf.WriteString("<p></p>")
			case "epigraph":
				buf.WriteString(`<blockquote class="epigraph">`)
			case "cite":
				buf.WriteString(`<blockquote class="cite">`)
			case "text-author":
				buf.WriteString(`<div class="text-author">`)
			case "v":
				buf.WriteString("<div>")
			case "stanza":
				buf.WriteString(`<div class="stanza">`)
			case "poem":
				buf.WriteString(`<div class="poem">`)
			case "subtitle":
				buf.WriteString(`<div class="subtitle">`)
			}
		case xml.EndElement:
			switch tok.Name.Local {
			case "annotation":
				return buf.String(), nil
			case "p":
				buf.WriteString("</p>")
			case "emphasis":
				buf.WriteString("</em>")
			case "strong":
				buf.WriteString("</strong>")
			case "strikethrough":
				buf.WriteString("</strike>")
			case "epigraph", "cite":
				buf.WriteString("</blockquote>")
			case "v", "stanza", "poem", "subtitle", "text-author":
				buf.WriteString("</div>")
			}
		case xml.CharData:
			buf.WriteString(html.EscapeString(string(tok)))
		}
	}
}

func parseCoverPage(d *xml.Decoder) (string, error) {
	for {
		tok, err := d.Token()
		if err != nil {
			return "", err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "image" {
				href := attr(tok, "href")
				if strings.HasPrefix(href, "#") {
					return href[1:], nil
				}
			}
		case xml.EndElement:
			if tok.Name.Local == "coverpage" {
				return "", nil
			}
		}
	}
}

func parseBinary(d *xml.Decoder) ([]byte, error) {
	var buf bytes.Buffer

	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}

		switch tok := tok.(type) {
		case xml.CharData:
			buf.Write(tok)
		case xml.EndElement:
			if tok.Name.Local == "binary" {
				return ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, &buf))
			}
		}
	}
}

func (b *book) makeImageName(href string) string {
	ext := strings.ToLower(path.Ext(href))
	sum := crc32.ChecksumIEEE([]byte(href))
	return fmt.Sprintf("%d_%d%s", b.ID, sum, ext)
}

func (b *book) AnnotationAndCover() (string, string, error) {
	r, err := b.OpenDeflate()
	if err != nil {
		return "", "", err
	}
	defer r.Close()

	d := xml.NewDecoder(r)
	d.CharsetReader = charset.NewReader

	var ann string
	var imageHref string
	var imageName string

	for {
		tok, err := d.Token()
		if err != nil {
			return "", "", err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "annotation":
				ann, err = parseAnnotation(d)
				if err != nil {
					return "", "", err
				}

			case "coverpage":
				if imageHref != "" {
					continue
				}

				imageHref, err = parseCoverPage(d)
				if err != nil {
					return "", "", err
				}
				if imageHref == "" {
					continue
				}

				imageName = b.makeImageName(imageHref)

				if p := imageCache.Get(imageName); p != nil {
					return ann, imageName, nil
				}
			case "body":
				err := skip(d, tok.Name)
				if err != nil {
					return "", "", err
				}
			case "binary":
				if id := attr(tok, "id"); id != imageHref {
					err := skip(d, tok.Name)
					if err != nil {
						return "", "", err
					}
					continue
				}

				data, err := parseBinary(d)
				if err != nil {
					return ann, "no-cover.png", nil
				}

				imageCache.Put(imageName, data)

				return ann, imageName, nil
			}
		case xml.EndElement:
			if tok.Name.Local == "FictionBook" {
				return ann, "", nil
			}
		}
	}
}

func (b *book) HTML() (string, error) {
	r, err := b.OpenDeflate()
	if err != nil {
		return "", err
	}
	defer r.Close()

	d := xml.NewDecoder(r)
	d.CharsetReader = charset.NewReader

	images := make(map[string]string)
	var buf bytes.Buffer

	for {
		tok, err := d.Token()
		if err != nil {
			return "", err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "description":
				err := skip(d, tok.Name)
				if err != nil {
					return "", err
				}
			case "body":
				buf.WriteString(`<div class="body">`)
			case "section":
				id := attr(tok, "id")
				if id != "" {
					buf.WriteString(`<a name="`)
					buf.WriteString(html.EscapeString(id))
					buf.WriteString(`"></a>`)
				}
				buf.WriteString(`<div class="section">`)
			case "title":
				buf.WriteString(`<div class="title">`)
			case "annotation":
				buf.WriteString(`<div class="annotation">`)
			case "text-author":
				buf.WriteString(`<div class="text-author">`)
			case "v":
				buf.WriteString("<div>")
			case "stanza":
				buf.WriteString(`<div class="stanza">`)
			case "poem":
				buf.WriteString(`<div class="poem">`)
			case "subtitle":
				buf.WriteString(`<div class="subtitle">`)
			case "p":
				buf.WriteString("<p>")
			case "emphasis":
				buf.WriteString("<em>")
			case "strong":
				buf.WriteString("<strong>")
			case "strikethrough":
				buf.WriteString("<strike>")
			case "empty-line":
				buf.WriteString("<p></p>")
			case "epigraph":
				buf.WriteString(`<blockquote class="epigraph">`)
			case "cite":
				buf.WriteString(`<blockquote class="cite">`)
			case "a":
				buf.WriteString("<a")
				href := attr(tok, "href")
				if strings.HasPrefix(href, "#") && attr(tok, "type") == "note" {
					buf.WriteString(` class="note" href="`)
					buf.WriteString(html.EscapeString(href))
					buf.WriteString(`"`)
				}
				buf.WriteString(">")
			case "image":
				href := attr(tok, "href")
				if strings.HasPrefix(href, "#") {
					imageName := b.makeImageName(href[1:])

					buf.WriteString(`<img src="/i/`)
					buf.WriteString(imageName)
					if alt := attr(tok, "alt"); alt != "" {
						buf.WriteString(`" alt="`)
						buf.WriteString(html.EscapeString(alt))
					}
					if title := attr(tok, "title"); title != "" {
						buf.WriteString(`" title="`)
						buf.WriteString(html.EscapeString(title))
					}
					buf.WriteString(`"/>`)

					images[href[1:]] = imageName
				}
			case "table":
				err := skip(d, tok.Name)
				if err != nil {
					return "", err
				}
			case "binary":
				if len(images) == 0 {
					return buf.String(), nil
				}

				id := attr(tok, "id")
				if imageName, ok := images[id]; ok {
					data, err := parseBinary(d)
					if err == nil {
						imageCache.Put(imageName, data)
					}
					delete(images, id)
				} else {
					err := skip(d, tok.Name)
					if err != nil {
						return "", err
					}
				}
			}

		case xml.EndElement:
			switch tok.Name.Local {
			case "body", "section", "title", "annotation", "text-author", "poem", "stanza", "v", "subtitle":
				buf.WriteString("</div>")
			case "epigraph", "cite":
				buf.WriteString("</blockquote>")
			case "p":
				buf.WriteString("</p>")
			case "emphasis":
				buf.WriteString("</em>")
			case "strong":
				buf.WriteString("</strong>")
			case "strikethrough":
				buf.WriteString("</strike>")
			case "a":
				buf.WriteString("</a>")
			case "FictionBook":
				return buf.String(), nil
			}

		case xml.CharData:
			buf.WriteString(html.EscapeString(string(tok)))
		}
	}
}

func (b *book) Image(sum uint32) ([]byte, error) {
	r, err := b.OpenDeflate()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	d := xml.NewDecoder(r)
	d.CharsetReader = charset.NewReader

	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "binary" {
				id := attr(tok, "id")
				if crc32.ChecksumIEEE([]byte(id)) == sum {
					return parseBinary(d)
				}

				err := skip(d, tok.Name)
				if err != nil {
					return nil, err
				}
			}

		case xml.EndElement:
			if tok.Name.Local == "FictionBook" {
				return nil, nil
			}
		}
	}
}
