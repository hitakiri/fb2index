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

func updateGenreWithBookCount(g *genre) error {
	var count int
	err := db.Get(&count, `SELECT COUNT(book_id)
				 FROM book_genres
				WHERE genre_id = ?
				`, g.ID)
	if err != nil {
		return err
	}

	g.BookCount = count

	return nil
}

func updateGenresWithBookCounts(genres []genre) error {
	for i := range genres {
		err := updateGenreWithBookCount(&genres[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func Genres() ([]genre, error) {
	var genres []genre
	err := db.Select(&genres, `SELECT id, name, desc, meta
				    FROM genres
				ORDER BY meta, desc, name`)
	if err != nil {
		return nil, err
	}

	err = updateGenresWithBookCounts(genres)
	if err != nil {
		return nil, err
	}

	return genres, nil
}
