// Copyright (C) 2023, VigilantDoomer
//
// This file is part of Offend program.
// It contains Sardinas-Patterson algorithm implementation:
// check if the dictionary can produce strictly non-ambiguous encodings,
// even if some words are prefixes of others
//
// Offend is free software: you can redistribute it
// and/or modify it under the terms of GNU General Public License
// as published by the Free Software Foundation, either version 3 of
// the License, or (at your option) any later version.
//
// Offend is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Offend.  If not, see <https://www.gnu.org/licenses/>.
package main

// TODO improve performance
import (
	"bytes"
	"sort"
)

func genSets(set [][]byte) [][][]byte {
	ret := [][][]byte{sortSet(set)}
	ret = append(ret, genFirstSet(ret[0]))
	// This may happen if all words are the same length,
	// or no prefix words already
	if len(ret[0]) == 0 {
		return ret
	}
	i := 2
	for {
		nSet := genNthSet(ret[0], ret[len(ret)-1])
		if len(nSet) == 0 {
			break
		}
		if setExists(nSet, ret) {
			break
		}
		i = i + 1
		ret = append(ret, nSet)
	}
	return ret
}

func setExists(set [][]byte, sets [][][]byte) bool {
	for _, v := range sets {
		if cmpSortedSet(set, v) {
			return true
		}
	}
	return false
}

func wordExists(wrd []byte, set [][]byte) bool {
	for _, v := range set {
		if bytes.Equal(v, wrd) {
			return true
		}
	}
	return false
}

// Returns true if two sets contain same elements
// The result is valid only for sorted sets
func cmpSortedSet(set1 [][]byte, set2 [][]byte) bool {
	if len(set1) != len(set2) {
		return false
	}
	for i, v := range set1 {
		if !bytes.Equal(v, set2[i]) {
			return false
		}
	}
	return true
}

// TODO refactor. GenFirstSet and GenNthSet don't need removal of duplicates,
// while the input set0 must be sorted and also not contain duplicate entries
// Thus it should be the caller's responsibility to do it
// TODO add removal of duplicates
// Too bad go doesn't provide it out the box even though the compare function
// is sufficient to implement the functionality with somekind of sort.SliceUniq
func sortSet(set [][]byte) [][]byte {
	ret := make([][]byte, len(set))
	for i := 0; i < len(set); i++ {
		ret[i] = set[i]
	}
	sort.Slice(ret, func(i, j int) bool {
		return bytes.Compare(ret[i], ret[j]) < 0
	})
	return ret
}

// Input parameter set0 must be sorted
// The returned set is sorted
func genFirstSet(set0 [][]byte) [][]byte {
	ret := makeCumul_Bytes2d()
	addSuffixesSecondNeedsToFormFirst(set0, set0, ret)
	return sortSet(*ret)
}

// Input parameters set0 and prevSet must be sorted
// The returned set is sorted
func genNthSet(set0 [][]byte, prevSet [][]byte) [][]byte {
	ret := makeCumul_Bytes2d()
	addSuffixesSecondNeedsToFormFirst(set0, prevSet, ret)
	addSuffixesSecondNeedsToFormFirst(prevSet, set0, ret)
	return sortSet(*ret)
}

func makeCumul_Bytes2d() *[][]byte {
	ret := make([][]byte, 0)
	return &ret
}

// Add suffixes that can be appended to words in set2
// to get words from set1. Cum accumulates the result
func addSuffixesSecondNeedsToFormFirst(set1 [][]byte, set2 [][]byte, cum *[][]byte) {
	tmp := *cum
	for _, u := range set1 {
		for _, v := range set2 {
			if bytes.HasPrefix(u, v) {
				// Fucking gotcha: the words may be equal, so there is no suffix to add
				wtf := u[len(v):]
				if len(wtf) > 0 && !wordExists(wtf, tmp) {
					tmp = append(tmp, wtf)
				}
			}
		}
	}
	*cum = tmp
}

func uniteSets(sets [][][]byte) [][]byte {
	ret := make([][]byte, 0)
	if len(sets) == 0 {
		return ret
	}
	for _, set := range sets {
		for _, wrd := range set {
			if !wordExists(wrd, ret) {
				ret = append(ret, wrd)
			}
		}
	}
	return sortSet(ret)
}

func hasIntersection(set1 [][]byte, set2 [][]byte) bool {
	// TODO optimize: make use of the fact that both are sorted
	for _, wrd := range set2 {
		if wordExists(wrd, set1) {
			return true
		}
	}
	return false
}

// Return true if no composition of words can have ambiguous decoding
// (true => dictionary is good for passphrase generation)
func SardinasPatterson_IsSafe(dict [][]byte) bool {
	CInfinity := genSets(dict)
	return !hasIntersection(dict, uniteSets(CInfinity[1:]))
}
