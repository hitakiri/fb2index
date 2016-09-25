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

func updateSequenceWithBookCount(s *sequence) error {
	var count int
	err := db.Get(&count, `SELECT COUNT(book_id)
				 FROM book_sequences
				WHERE sequence_id = ?
				`, s.ID)
	if err != nil {
		return err
	}

	s.BookCount = count

	return nil
}

func updateSequencesWithBookCounts(sequences []sequence) error {
	for i := range sequences {
		err := updateSequenceWithBookCount(&sequences[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func SequencesPerPage(n int) ([]sequence, int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM sequences")
	if err != nil {
		return nil, 0, err
	}

	numPages := (count + *sequencesPerPage - 1) / *sequencesPerPage
	offset := (n - 1) * *sequencesPerPage
	if offset >= count {
		return nil, numPages, nil
	}

	var sequences []sequence
	err = db.Select(&sequences, `SELECT id, name
				       FROM sequences
				   ORDER BY name
				      LIMIT ?, ?
				`, offset, *sequencesPerPage)
	if err != nil {
		return nil, 0, err
	}

	err = updateSequencesWithBookCounts(sequences)
	if err != nil {
		return nil, 0, err
	}

	return sequences, numPages, nil
}

func SequenceByID(id uint32) (*sequence, error) {
	var s sequence
	err := db.Get(&s, "SELECT id, name FROM sequences WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
