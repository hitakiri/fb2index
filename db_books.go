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

import "sort"

func bookGenres(id uint32) ([]genre, error) {
	var ge []genre
	err := db.Select(&ge, `SELECT id, name, desc
				 FROM book_genres bg, genres g
				WHERE bg.genre_id = g.id
				  AND bg.book_id = ?
			     ORDER BY desc
				`, id)
	return ge, err
}

func bookAuthors(id uint32) ([]author, error) {
	var au []author
	err := db.Select(&au, `SELECT id, first_name, middle_name, last_name, nickname
				 FROM book_authors ba, authors a
				WHERE ba.author_id = a.id
				  AND ba.book_id = ?
			     ORDER BY last_name, first_name, nickname
				`, id)
	return au, err
}

func bookTranslators(id uint32) ([]author, error) {
	var tr []author
	err := db.Select(&tr, `SELECT id, first_name, middle_name, last_name, nickname
				 FROM book_translators bt, authors a
				WHERE bt.author_id = a.id
				  AND bt.book_id = ?
			     ORDER BY last_name, first_name, nickname
				`, id)
	return tr, err
}

func bookSequences(id uint32) ([]sequence, error) {
	var sq []sequence
	err := db.Select(&sq, `SELECT id, name, number
				 FROM book_sequences bs, sequences s
				WHERE bs.sequence_id = s.id
				  AND bs.book_id = ?
			     ORDER BY name, number
				`, id)
	return sq, err
}

func (b *book) FetchRelations() error {
	genres, err := bookGenres(b.ID)
	if err != nil {
		return err
	}
	b.Genres = genres

	authors, err := bookAuthors(b.ID)
	if err != nil {
		return err
	}
	b.Authors = authors

	translators, err := bookTranslators(b.ID)
	if err != nil {
		return err
	}
	b.Translators = translators

	sequences, err := bookSequences(b.ID)
	if err != nil {
		return err
	}
	b.Sequences = sequences

	return nil
}

func fetchBooksRelations(books []book) error {
	for i := range books {
		err := books[i].FetchRelations()
		if err != nil {
			return err
		}
	}

	return nil
}

func BooksPerPage(n int) ([]book, int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM books")
	if err != nil {
		return nil, 0, err
	}

	numPages := (count + *booksPerPage - 1) / *booksPerPage
	offset := (n - 1) * *booksPerPage
	if offset >= count {
		return nil, numPages, nil
	}

	var books []book
	err = db.Select(&books, `SELECT id, title, lang, archive, filename, offset, compressed_size, uncompressed_size, crc32
				   FROM books
			       ORDER BY title
				  LIMIT ?, ?
				`, offset, *booksPerPage)
	if err != nil {
		return nil, 0, err
	}

	err = fetchBooksRelations(books)
	if err != nil {
		return nil, 0, err
	}

	return books, numPages, nil
}

func BooksByGenre(id uint32) ([]book, *genre, error) {
	var g genre
	err := db.Get(&g, "SELECT id, name, desc FROM genres WHERE id = ?", id)
	if err != nil {
		return nil, nil, err
	}

	var books []book
	err = db.Select(&books, `SELECT id, title, lang, archive, filename, offset, compressed_size, uncompressed_size, crc32
				   FROM books b, book_genres bg
				  WHERE b.id = bg.book_id
				    AND bg.genre_id = ?
			       ORDER BY title
				`, id)
	if err != nil {
		return nil, nil, err
	}

	err = fetchBooksRelations(books)
	if err != nil {
		return nil, nil, err
	}

	return books, &g, nil
}

func BooksByAuthor(id uint32) ([]book, []book, *author, error) {
	var a author
	err := db.Get(&a, `SELECT id, first_name, middle_name, last_name, nickname
			     FROM authors
			    WHERE id = ?
			    `, id)
	if err != nil {
		return nil, nil, nil, err
	}

	var books []book
	err = db.Select(&books, `SELECT id, title, lang, archive, filename, offset, compressed_size, uncompressed_size, crc32
				   FROM books b, book_authors ba
				  WHERE b.id = ba.book_id
				    AND ba.author_id = ?
			       ORDER BY title
				`, id)
	if err != nil {
		return nil, nil, nil, err
	}

	err = fetchBooksRelations(books)
	if err != nil {
		return nil, nil, nil, err
	}

	var translations []book
	err = db.Select(&translations, `SELECT id, title, lang, archive, filename, offset, compressed_size, uncompressed_size, crc32
					  FROM books b, book_translators bt
					 WHERE b.id = bt.book_id
					   AND bt.author_id = ?
				      ORDER BY title
				`, id)
	if err != nil {
		return nil, nil, nil, err
	}

	err = fetchBooksRelations(translations)
	if err != nil {
		return nil, nil, nil, err
	}

	return books, translations, &a, nil
}

type booksBySequence []book

func (b booksBySequence) Len() int { return len(b) }
func (b booksBySequence) Less(i, j int) bool {
	return b[i].Sequences[0].Number < b[j].Sequences[0].Number
}
func (b booksBySequence) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func BooksBySequence(id uint32) ([]book, *sequence, error) {
	var seq sequence
	err := db.Get(&seq, "SELECT id, name FROM sequences WHERE id = ?", id)
	if err != nil {
		return nil, nil, err
	}

	var books []book
	err = db.Select(&books, `SELECT id, title, lang, archive, filename, offset, compressed_size, uncompressed_size, crc32
				   FROM books b, book_sequences bs
				  WHERE b.id = bs.book_id
				    AND bs.sequence_id = ?
			       ORDER BY number, title
				`, id)
	if err != nil {
		return nil, nil, err
	}

	err = fetchBooksRelations(books)
	if err != nil {
		return nil, nil, err
	}

	for i := range books {
		sequences := books[i].Sequences
		for j := 1; j < len(sequences); j++ {
			if sequences[j].ID == id {
				sequences[0], sequences[j] = sequences[j], sequences[0]
				break
			}
		}
	}
	sort.Sort(booksBySequence(books))

	return books, &seq, nil
}

func BookByID(id uint32) (*book, error) {
	var b book
	err := db.Get(&b, `SELECT id, title, lang, archive, filename, offset, compressed_size, uncompressed_size, crc32
			     FROM books
			    WHERE id = ?
				`, id)

	if err != nil {
		return nil, err
	}

	err = b.FetchRelations()
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func BookByIDWithAnnotation(id uint32) (*book, string, string, error) {
	b, err := BookByID(id)
	if err != nil {
		return nil, "", "", err
	}

	ann, cover, err := b.AnnotationAndCover()
	if err != nil {
		return nil, "", "", err
	}

	return b, ann, cover, nil
}
