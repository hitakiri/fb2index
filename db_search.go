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

import "github.com/opennota/fb2index/trigram"

var (
	trgmAuthorIndex   = trigram.NewIndex()
	trgmSequenceIndex = trigram.NewIndex()
	trgmBookIndex     = trigram.NewIndex()
)

func Search(query string) (authors []author, sequences []sequence, books []book, err error) {
	trgm := trigram.Extract(query)
	authorIDs := trgmAuthorIndex.QueryTrigrams(trgm)
	sequenceIDs := trgmSequenceIndex.QueryTrigrams(trgm)
	bookIDs := trgmBookIndex.QueryTrigrams(trgm)

	if len(authorIDs) > 0 {
		authors = make([]author, 0, len(authorIDs))
	}
	if len(sequenceIDs) > 0 {
		sequences = make([]sequence, 0, len(sequenceIDs))
	}
	if len(bookIDs) > 0 {
		books = make([]book, 0, len(bookIDs))
	}

	for _, id := range authorIDs {
		a, err := AuthorByID(id)
		if err != nil {
			return nil, nil, nil, err
		}

		authors = append(authors, *a)
	}

	for _, id := range sequenceIDs {
		s, err := SequenceByID(id)
		if err != nil {
			return nil, nil, nil, err
		}

		sequences = append(sequences, *s)
	}

	for _, id := range bookIDs {
		b, err := BookByID(id)
		if err != nil {
			return nil, nil, nil, err
		}

		books = append(books, *b)
	}

	return authors, sequences, books, nil
}
