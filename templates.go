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
)

var (
	templates = make(map[string]*template.Template)

	templateSources = map[string]string{
		"book_index":     bookIndexTmpl,
		"genre_index":    genreIndexTmpl,
		"author_index":   authorIndexTmpl,
		"sequence_index": sequenceIndexTmpl,
		"book":           bookTmpl,
		"book_read":      bookReadTmpl,
		"genre":          genreTmpl,
		"author":         authorTmpl,
		"sequence":       sequenceTmpl,
		"search":         searchTmpl,
	}
)

var funcs = template.FuncMap{
	"inc": func(n int) int { return n + 1 },
	"dec": func(n int) int { return n - 1 },
	"prmeta": func(genres []genre, index int) string {
		if index == 0 {
			return ""
		}
		return genres[index-1].Meta
	},
	"hrsize": func(size int64) string {
		switch {
		case size > 1073741824:
			return fmt.Sprintf("%.1f Гб", float64(size)/1073741824)
		case size > 1048576:
			return fmt.Sprintf("%.1f Мб", float64(size)/1048576)
		case size > 1024:
			return fmt.Sprintf("%d Кб", size/1024)
		default:
			return fmt.Sprintf("%d б", size)
		}
	},
}

func mustParse(data ...string) *template.Template {
	var root *template.Template
	for _, s := range data {
		var t *template.Template
		if root == nil {
			root = template.New("")
			t = root
		} else {
			t = root.New("")
		}
		_, err := t.Funcs(funcs).Parse(s)
		if err != nil {
			panic(err)
		}
	}
	return root
}

func init() {
	for name, source := range templateSources {
		if name == "book_index" || name == "author_index" || name == "sequence_index" {
			templates[name] = mustParse(source, pager, base)
		} else {
			templates[name] = mustParse(source, base)
		}
	}
}

var base = `
{{ define "base" }}
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ template "title" . }}</title>
  <style>
    body {
      margin: 40px auto;
      max-width: 650px;
      line-height: 1.6;
      font-size: 18px;
      color: #444;
      padding: 0 10px;
    }
    h1, h2, h3 {
      line-height: 1.2;
    }
    .book {
      margin-bottom: 10px;
    }
    .book:nth-child(even) {
      background-color: #ddd;
    }
    .book-lang,
    .book-genres,
    .book-authors,
    .book-translators,
    .book-sequences {
      font-size: small;
      color: #aaa;
    }
    .book-genre:not(:last-child):after,
    .book-author:not(:last-child):after,
    .book-translator:not(:last-child):after,
    .book-sequence:not(:last-child):after {
      content: ",";
    }
    .text-pages {
      color: #aaa;
    }
    .num-books {
      font-size: small;
      color: #aaa;
    }
    .author:nth-child(even) {
      background-color: #ddd;
    }
    .sequence:nth-child(even) {
      background-color: #ddd;
    }
    .genre {
      margin-left: 20px;
    }
    {{ template "styles" }}
  </style>
  <link rel="stylesheet" type="text/css" href="/external.css">
</head>
<body>
  <header>
    <nav>
      <a class="top-nav-link" href="/b">Книги</a>
      <a class="top-nav-link" href="/g">Жанры</a>
      <a class="top-nav-link" href="/a">Авторы</a>
      <a class="top-nav-link" href="/s">Серии</a>
      <a class="top-nav-link" href="/search">Поиск</a>
    </nav>
    <h1>{{ template "title" . }}</h1>
  </header>
  <nav>
    {{ template "pager" . }}
  </nav>
  <main>
    {{ template "main" . }}
  </main>
  <nav>
    {{ template "pager" . }}
  </nav>
</body>
</html>
{{ end }}
{{ define "book_genres" }}
  {{ if .Genres }}
    <div class="book-genres">
      <span class="text-genres">Жанр{{ if gt (len .Genres) 1 }}ы{{ end }}:</span>
        {{ range .Genres }}
          <span class="book-genre">
            <a class="genre-link" href="/g?id={{ .ID }}">
              {{ if .Desc }}{{ .Desc }}{{ else }}{{ .Name }}{{ end }}</a>
          </span>
        {{ end }}
    </div>
  {{ end }}
{{ end }}
{{ define "book_authors" }}
  {{ if .Authors }}
    <div class="book-authors">
        <span class="text-authors">Автор{{ if gt (len .Authors) 1 }}ы{{ end }}:</span>
        {{ range .Authors }}
          <span class="book-author">
            <a class="author-link" href="/a?id={{ .ID }}">
              {{ .FirstName }} {{ .MiddleName }} {{ .LastName }}{{ if .Nickname }}{{ if or .FirstName .LastName }} (aka {{ .Nickname }}){{ else }}{{ .Nickname }}{{ end }}{{ end }}</a>
          </span>
        {{ end }}
    </div>
  {{ end }}
{{ end }}
{{ define "book_translators" }}
  {{ if .Translators }}
    <div class="book-translators">
        <span class="text-translators">Переводчик{{ if gt (len .Translators) 1 }}и{{ end }}:</span>
        {{ range .Translators }}
          <span class="book-translator">
            <a class="author-link" href="/a?id={{ .ID }}">
              {{ .FirstName }} {{ .MiddleName }} {{ .LastName }}{{ if .Nickname }}{{ if or .FirstName .LastName }} (aka {{ .Nickname }}){{ else }}{{ .Nickname }}{{ end }}{{ end }}</a>
          </span>
        {{ end }}
    </div>
  {{ end }}
{{ end }}
{{ define "book_sequences" }}
  {{ if .Sequences }}
    <div class="book-sequences">
      <span class="text-series">Сери{{ if gt (len .Sequences) 1 }}и{{ else }}я{{ end }}:</span>
        {{ range .Sequences }}
          <span class="book-sequence">
            <a class="sequence-link" href="/s?id={{ .ID }}">{{ .Name }}</a>{{ if .Number }}-{{ .Number }}{{ end }}
          </span>
        {{ end }}
    </div>
  {{ end }}
{{ end }}
{{ define "book_count" }}
  <div class="num-books">
    <span class="text-num-books">Книг:</span>
    <span class="book-count">{{ .BookCount }}</span>
  </div>
{{ end }}
{{ define "pager" }}{{ end }}
{{ define "styles" }}{{ end }}
`

var pager = `
{{ define "pager" }}
  {{ if gt .TotalPages 1 }}
    {{ $PrevPage := dec .PageNumber }}
    {{ $NextPage := inc .PageNumber }}
    <span class="text-pages">Страницы:</span>
    {{ if gt $PrevPage 1 }}
      <a class="first-page-link" href="/{{ template "prefix" }}">{{ if gt (dec $PrevPage) 1 }}Первая{{ else }}1{{ end }}</a>
    {{ end }}
    {{ if gt (dec $PrevPage) 1 }}...{{ end }}
    {{ if ge $PrevPage 1 }}
      <a class="prev-page-link" href="/{{ template "prefix" }}?page={{ $PrevPage }}">{{ $PrevPage }}</a>
    {{ end }}
    <span class="current-page-number">{{ .PageNumber }}</span>
    {{ if le $NextPage .TotalPages }}
      <a class="next-page-link" href="/{{ template "prefix" }}?page={{ $NextPage }}">{{ $NextPage }}</a>
    {{ end }}
    {{ if lt (inc $NextPage) .TotalPages }}...{{ end }}
    {{ if lt $NextPage .TotalPages }}
    <a class="last-page-link" href="/{{ template "prefix" }}?page={{ .TotalPages }}">{{ if lt (inc $NextPage) .TotalPages }}Последняя{{ else }}{{ .TotalPages }}{{ end }}</a>
    {{ end }}
  {{ end }}
{{ end }}
{{ define "prefix" }}{{ end }}
`

var bookIndexTmpl = `
{{ define "prefix" }}b{{ end }}
{{ define "title" }}Книги{{ end }}
{{ define "main" }}
  {{ range .Books }}
    <div class="book">
      <div class="book-title">
        <a class="book-link" href="/b?id={{ .ID }}">{{ .Title }}</a>
      </div>
      {{ template "book_genres" . }}
      {{ template "book_authors" . }}
      {{ template "book_translators" . }}
      {{ template "book_sequences" . }}
    </div>
  {{ end }}
{{ end }}
`

var genreIndexTmpl = `
{{ define "title" }}Жанры{{ end }}
{{ define "main" }}
  {{ range $i, $g := . }}
    {{ if ne .Meta (prmeta $ $i) }}
      <div class="genre-meta">
        <h2>{{ .Meta }}</h2>
      </div>
    {{ end }}
    <div class="genre">
      <div class="genre-name">
        <a class="genre-link" href="/g?id={{ .ID }}">
          <span class="genre-desc">{{ if .Desc }}{{ .Desc }}{{ else }}{{ .Name }}{{ end }}</span></a>
        {{ if .Desc }}(<span class="genre-name">{{ .Name }}</span>){{ end }}
      </div>
      {{ template "book_count" . }}
    </div>
  {{ end }}
{{ end }}
`

var authorIndexTmpl = `
{{ define "prefix" }}a{{ end }}
{{ define "title"  }}Авторы{{ end }}
{{ define "main" }}
  {{ range .Authors }}
    <div class="author">
      <div class="author-name">
        <a class="author-link" href="/a?id={{ .ID }}">
          <span class="first-name">{{ .FirstName }}</span>
          <span class="middle-name">{{ .MiddleName }}</span>
          <span class="last-name">{{ .LastName }}</span>
          {{ if .Nickname }}
            {{ if or .FirstName .MiddleName .LastName }}aka{{ end }} <span class="nickname">{{ .Nickname }}</span>
          {{ end }}
        </a>
      </div>
      {{ template "book_count" . }}
    </div>
  {{ end }}
{{ end }}
`

var sequenceIndexTmpl = `
{{ define "prefix" }}s{{ end }}
{{ define "title" }}Серии{{ end }}
{{ define "main" }}
  {{ range .Sequences }}
    <div class="sequence">
      <div class="sequence-name">
        <a class="sequence-link" href="/s?id={{ .ID }}">
          <span class="seq-name">{{ .Name }}</span>
        </a>
      </div>
      {{ template "book_count" . }}
    </div>
  {{ end }}
{{ end }}
`

var genreTmpl = `
{{ define "title" }}
  Жанры / {{ if .Genre.Desc }}{{ .Genre.Desc }}{{ else }}{{ .Genre.Name }}{{ end }}
{{ end }}
{{ define "main" }}
  {{ range .Books }}
    <div class="book">
      <div class="book-title">
        <a class="book-link" href="/b?id={{ .ID }}">{{ .Title }}</a>
      </div>
      {{ if gt (len .Genres) 1 }}
        <div class="book-genres">
          <span class="text-genres">Жанр{{ if gt (len .Genres) 2 }}ы{{ end }}:</span>
            {{ range .Genres }}
              {{ if ne .ID $.Genre.ID }}
                <span class="book-genre">
                  <a class="genre-link" href="/g?id={{ .ID }}">
                    {{ if .Desc }}{{ .Desc }}{{ else }}{{ .Name }}{{ end }}</a>
                </span>
              {{ end }}
            {{ end }}
        </div>
      {{ end }}
      {{ template "book_authors" . }}
      {{ template "book_translators" . }}
      {{ template "book_sequences" . }}
    </div>
  {{ end }}
{{ end }}
`

var authorTmpl = `
{{ define "title" }}
  Авторы / {{ .Author.FirstName }} {{ .Author.MiddleName }} {{ .Author.LastName }}
{{ end }}
{{ define "main" }}
  <div class="author-books">
    {{ range .Books }}
      <div class="book">
        <div class="book-title">
          <a class="book-link" href="/b?id={{ .ID }}">{{ .Title }}</a>
        </div>
        {{ template "book_genres" . }}
        {{ if gt (len .Authors) 1 }}
          <div class="book-authors">
              <span class="text-authors">Соавтор{{ if gt (len .Authors) 2 }}ы{{ end }}:</span>
              {{ range .Authors }}
                {{ if ne .ID $.Author.ID }}
                  <span class="book-author">
                    <a class="author-link" href="/a?id={{ .ID }}">
                      {{ .FirstName }} {{ .MiddleName }} {{ .LastName }}{{ if .Nickname }}{{ if or .FirstName .LastName }} (aka {{ .Nickname }}){{ else }}{{ .Nickname }}{{ end }}{{ end }}</a>
                  </span>
                {{ end }}
              {{ end }}
          </div>
        {{ end }}
        {{ template "book_translators" . }}
        {{ template "book_sequences" . }}
      </div>
    {{ end }}
  </div>
  {{ if .Translations }}
    <h2>Переводы</h2>
    <div class="author-translations">
      {{ range .Translations }}
        <div class="book">
          <div class="book-title">
            <a class="book-link" href="/b?id={{ .ID }}">{{ .Title }}</a>
          </div>
          {{ template "book_genres" . }}
          {{ template "book_authors" . }}
          {{ if gt (len .Translators) 1 }}
            <div class="book-translators">
              <span class="text-translators">Сопереводчик{{ if gt (len .Translators) 2 }}и{{ end }}:</span>
                {{ range .Translators }}
                  {{ if ne .ID $.Author.ID }}
                    <span class="book-translator">
                      <a class="author-link" href="/a?id={{ .ID }}">
                        {{ .FirstName }} {{ .MiddleName }} {{ .LastName }}{{ if .Nickname }}{{ if or .FirstName .LastName }} (aka {{ .Nickname }}){{ else }}{{ .Nickname }}{{ end }}{{ end }}</a>
                    </span>
                  {{ end }}
                {{ end }}
            </div>
          {{ end }}
          {{ template "book_sequences" . }}
        </div>
      {{ end }}
    </div>
  {{ end }}
{{ end }}
`

var sequenceTmpl = `
{{ define "title" }}
  Серии / {{ .Sequence.Name }}
{{ end }}
{{ define "main" }}
  {{ range .Books }}
    <div class="book">
      <div class="book-title">
        <span class="number-in-sequence">
          {{ with $n := (index .Sequences 0).Number }}
            {{ if $n }}{{ $n }}.{{ end }} -
          {{ end }}
        </span>
        <a class="book-link" href="/b?id={{ .ID }}">{{ .Title }}</a>
      </div>
      {{ template "book_genres" . }}
      {{ template "book_authors" . }}
      {{ template "book_translators" . }}
      {{ if gt (len .Sequences) 1 }}
        <div class="book-sequences">
          <span class="text-series">Сери{{ if gt (len .Sequences) 2 }}и{{ else }}я{{ end }}:</span>
            {{ range .Sequences }}
              {{ if ne .ID ($.Sequence.ID) }}
                <span class="book-sequence">
                  <a class="sequence-link" href="/s?id={{ .ID }}">{{ .Name }}</a>{{ if .Number }}-{{ .Number }}{{ end }}
                </span>
              {{ end }}
            {{ end }}
        </div>
      {{ end }}
    </div>
  {{ end }}
{{ end }}
`

var bookTmpl = `
{{ define "title" }}
  Книги / {{ .Book.Title }}
{{ end }}
{{ define "styles" }}
  .buttons {
    float: right;
  }
  .annotation > img {
    max-width: 250px;
    max-height: 300px;
    float: left;
    margin: 10px;
  }
{{ end }}
{{ define "main" }}
  <div class="book-lang">
    <span class="book-lang-text">Язык:</span>
    {{ if .Language }}
      <a href="#">{{ .Language }}</a>
    {{ else }}
      {{ .Book.Lang }}
    {{ end }}
  </div>
  {{ with .Book }}
    {{ template "book_genres" . }}
    {{ template "book_authors" . }}
    {{ template "book_translators" . }}
    {{ template "book_sequences" . }}
  {{ end }}
  <div class="buttons">
    <a class="read-button" href="/b?id={{ .Book.ID }}&action=read">
      (читать)</a>
    <a class="download-button" href="/b?id={{ .Book.ID }}&action=download">
      (скачать ({{ hrsize .Book.UncompressedSize }}))
    </a>
  </div>
  <div class="annotation">
    <img class="book-cover" src="/i/{{ if .Cover }}{{ .Cover }}{{ else }}no-cover.png{{ end }}">
    {{ .Ann }}
  </div>
{{ end }}
`

var bookReadTmpl = `
{{ define "title" }}
  Книги / {{ .Book.Title }}
{{ end }}
{{ define "styles" }}
  .title {
    margin-left: 0;
    margin-right: 0;
    font-weight: bold;
  }
  .body > .title {
    font-size: 1.5em;
    margin-top: 0.83em;
    margin-bottom: 0.83em;
  }
  .section > .title {
    font-size: 1.17em;
    margin-top: 1em;
    margin-bottom: 1em;
  }
  .section > .section > .title {
    margin-top: 1.33em;
    margin-bottom: 1.33em;
  }
  .section > .section > .section > .title {
    font-size: .83em;
    margin-top: 1.67em;
    margin-bottom: 1.67em;
  }
  .title {
    font-size: .67em;
    margin-top: 2.33em;
    margin-bottom: 2.33em;
  }
  .text-author {
    text-align: right;
  }
  .subtitle {
    text-align: center;
    font-weight: bold;
  }
  .note {
    vertical-align: super;
  }
  .book-link {
    font-size: small;
  }
{{ end }}
{{ define "main" }}
  <div>
    <a class="book-link" href="/b?id={{ .Book.ID }}">Страница книги</a>
  </div>
  <div class="content">
    {{ .HTML }}
  </div>
{{ end }}
`

var searchTmpl = `
{{ define "title" }}Поиск{{ end }}
{{ define "main" }}
  <div class="search-form">
    <form method="POST" action="/search">
      <div>
        <input type="text" name="query" value="{{ .SearchQuery }}">
        <button type="submit">Искать</button>
      </div>
    </form>
  </div>
  {{ if .SearchQuery }}
    <div class="search-results">
      {{ if not (or .Authors .Sequences .Books) }}Ничего не найдено.{{ end }}
      {{ if .Authors }}
        <h2>Найденные авторы</h2>
        <div class="search-results-authors">
          {{ range .Authors }}
            <div class="author">
              <a class="author-link" href="/a?id={{ .ID }}">
                {{ .FirstName }} {{ .MiddleName }} {{ .LastName }}{{ if .Nickname }}{{ if or .FirstName .LastName }} (aka {{ .Nickname }}){{ else }}{{ .Nickname }}{{ end }}{{ end }}</a>
            </div>
          {{ end }}
        </div>
      {{ end }}
      {{ if .Sequences }}
        <h2>Найденные серии</h2>
        <div class="search-results-sequences">
          {{ range .Sequences }}
            <div class="sequence">
              <a class="sequence-link" href="/s?id={{ .ID }}">{{ .Name }}</a>
            </div>
          {{ end }}
        </div>
      {{ end }}
      {{ if .Books }}
        <h2>Найденные книги</h2>
        <div class="search-results-books">
          {{ range .Books }}
            <div class="book">
              <div class="book-title">
                <a class="book-link" href="/b?id={{ .ID }}">{{ .Title }}</a>
              </div>
              {{ template "book_genres" . }}
              {{ template "book_authors" . }}
              {{ template "book_translators" . }}
              {{ template "book_sequences" . }}
            </div>
          {{ end }}
        </div>
      {{ end }}
    </div>
  {{ end }}
{{ end }}
`
