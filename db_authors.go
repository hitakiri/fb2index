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

func updateAuthorWithBookCount(a *author) error {
	var count int
	err := db.Get(&count, `SELECT COUNT(book_id) FROM (
					SELECT book_id, author_id
					  FROM book_authors
					 WHERE author_id = $1
					 UNION
					SELECT book_id, author_id
					  FROM book_translators
					 WHERE author_id = $1
				)
				`, a.ID)
	if err != nil {
		return err
	}

	a.BookCount = count

	return nil
}

func updateAuthorsWithBookCounts(authors []author) error {
	for i := range authors {
		err := updateAuthorWithBookCount(&authors[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func AuthorsPerPage(n int) ([]author, int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM authors")
	if err != nil {
		return nil, 0, err
	}

	numPages := (count + *authorsPerPage - 1) / *authorsPerPage
	offset := (n - 1) * *authorsPerPage
	if offset >= count {
		return nil, numPages, nil
	}

	var authors []author
	err = db.Select(&authors, `SELECT id, first_name, middle_name, last_name, nickname
				     FROM authors
				 ORDER BY last_name, first_name, nickname
				    LIMIT ?, ?
				`, offset, *authorsPerPage)
	if err != nil {
		return nil, 0, err
	}

	err = updateAuthorsWithBookCounts(authors)
	if err != nil {
		return nil, 0, err
	}

	return authors, numPages, nil
}

func AuthorByID(id uint32) (*author, error) {
	var au author
	err := db.Get(&au, `SELECT id, first_name, middle_name, last_name, nickname
				 FROM authors
				WHERE id = ?
				`, id)
	if err != nil {
		return nil, err
	}

	return &au, nil
}
