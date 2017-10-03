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

// Package trigram is a simple trigram index.
package trigram

import (
	"hash/crc32"
	"sort"
	"unicode"
	"unicode/utf8"
)

// T is a trigram.
type T uint32

const tAllIDs T = 0xffffffff

// Index is a trigram index.
type Index map[T][]uint32

// NewIndex returns a new trigram index.
func NewIndex() Index {
	return make(Index)
}

func mkT(rr [3]rune) T {
	var p [12]byte
	n := utf8.EncodeRune(p[0:], rr[0])
	n += utf8.EncodeRune(p[n:], rr[1])
	n += utf8.EncodeRune(p[n:], rr[2])
	return T((crc32.ChecksumIEEE(p[:n]) >> 6) & 0xffffff)
}

func normalize(r rune) rune {
	if r == 'ё' {
		return 'е'
	}
	return r
}

func appendUnique(tt []T, t T) []T {
	for _, v := range tt {
		if v == t {
			return tt
		}
	}
	return append(tt, t)
}

// Extract returns a slice of all the unique trigrams in s.
func Extract(s string) []T {
	if s == "" {
		return nil
	}

	rr := [3]rune{' ', ' ', ' '}
	tt := make([]T, 0, len(s))
	i := 0

	for {
		r, size := utf8.DecodeRuneInString(s[i:])
		if size == 0 {
			if rr[1] != ' ' {
				rr[2] = ' '
				tt = appendUnique(tt, mkT(rr))
			}
			break
		}
		i += size

		if unicode.IsLetter(r) {
			rr[2] = normalize(unicode.ToLower(r))
		} else if unicode.IsDigit(r) {
			rr[2] = r
		} else if rr[1] != ' ' && unicode.IsSpace(r) {
			rr[2] = ' '
		} else {
			continue
		}

		tt = appendUnique(tt, mkT(rr))

		rr[0] = rr[1]
		rr[1] = rr[2]
	}

	if len(tt) == 0 {
		return nil
	}

	return tt
}

// Add adds a string under the given ID.
func (idx Index) Add(id uint32, s string) {
	idx.AddTrigrams(id, Extract(s))
}

// AddTrigrams adds a slice of trigrams under the given ID.
func (idx Index) AddTrigrams(id uint32, tt []T) {
	if len(tt) == 0 {
		return
	}

	allIDs := idx[tAllIDs]
	l := len(allIDs)
	if l > 0 && allIDs[l-1] > id {
		panic("out of order")
	}
	if l == 0 || allIDs[l-1] != id {
		idx[tAllIDs] = append(allIDs, id)
	}

	for _, t := range tt {
		idxt := idx[t]
		l := len(idxt)
		if l == 0 || idxt[l-1] != id {
			idx[t] = append(idxt, id)
		}
	}
}

// Query returns a slice of IDs that match the trigrams in the query s.
func (idx Index) Query(s string) []uint32 {
	return idx.QueryTrigrams(Extract(s))
}

type byRelevance struct {
	ids []uint32
	rel []uint32
}

func (p byRelevance) Len() int           { return len(p.ids) }
func (p byRelevance) Less(i, j int) bool { return p.rel[i] > p.rel[j] }
func (p byRelevance) Swap(i, j int) {
	p.ids[i], p.ids[j] = p.ids[j], p.ids[i]
	p.rel[i], p.rel[j] = p.rel[j], p.rel[i]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// QueryTrigrams returns a slice of IDs that match the given set of trigrams.
func (idx Index) QueryTrigrams(tt []T) []uint32 {
	if len(tt) == 0 {
		return nil
	}

	l := 0
	for _, t := range tt {
		l += len(idx[t])
	}
	l = min(l, len(idx[tAllIDs]))

	m := make(map[uint32]uint32, l)
	for _, t := range tt {
		for _, id := range idx[t] {
			m[id]++
		}
	}

	threshold := uint32(float64(len(tt)) * 0.75)
	ids := make([]uint32, 0, len(m))
	rel := make([]uint32, 0, len(m))
	for id := range m {
		r := m[id]
		if r >= threshold {
			ids = append(ids, id)
			rel = append(rel, r)
		}
	}

	sort.Sort(byRelevance{ids, rel})

	return ids
}
