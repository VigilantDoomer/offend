// Copyright (C) 2023, VigilantDoomer
//
// This file is part of Offend program.
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

import (
	"testing"
)

type analyzer_testrecord struct {
	input          []string
	distinctCounts [][]int
	prefixExists   bool
}

func produceDupTracker(input []string) map[string]int {
	dupTracker := make(map[string]int)
	for _, wrd := range input {
		oldCnt, _ := dupTracker[wrd]
		dupTracker[wrd] = oldCnt + 1
	}
	return dupTracker
}

func cmpDistinctCounts(d1 [][]int, d2 [][]int) bool {
	// Must be sorted by d_[i][0] and d_[i][0] must have unique values
	if len(d1) != len(d2) {
		return false
	}
	for i, d1i := range d1 {
		if (d1i[0] != d2[i][0]) || (d1i[1] != d2[i][1]) {
			return false
		}
	}
	return true
}

func TestGetDistinctCountsAndDoPrefixCheck(t *testing.T) {
	dataset := []analyzer_testrecord{
		analyzer_testrecord{input: []string{"dog", "cat", "doggerel"}, distinctCounts: [][]int{{1, 3}}, prefixExists: true},
		analyzer_testrecord{input: []string{"dog", "dog", "doggerel", "dog", "doggerel"}, distinctCounts: [][]int{{2, 1}, {3, 1}}, prefixExists: true},
		analyzer_testrecord{input: []string{"dog", "dog", "doggerel", "dog", "doggerel", "doggerel"}, distinctCounts: [][]int{{3, 2}}, prefixExists: true},
		analyzer_testrecord{input: []string{"dog", "dog", "cat", "dog", "cat", "cat"}, distinctCounts: [][]int{{3, 2}}, prefixExists: false},
		analyzer_testrecord{input: []string{"dog", "cat", "dogcat"}, distinctCounts: [][]int{{1, 3}}, prefixExists: true},
		analyzer_testrecord{input: []string{"dog", "cat", "donut", "nutjob", "nutcat", "catdoggery"}, distinctCounts: [][]int{{1, 6}}, prefixExists: true},
		analyzer_testrecord{input: []string{"dog", "cat", "doggerel", "gerelcat"}, distinctCounts: [][]int{{1, 4}}, prefixExists: true},
		analyzer_testrecord{input: []string{"doggie", "catmonster", "doggerel", "doggy"}, distinctCounts: [][]int{{1, 4}}, prefixExists: false},
	}
	for num, testrecord := range dataset {
		dupTracker := produceDupTracker(testrecord.input)
		distinctCounts, prefixExample := getDistinctCountsAndDoPrefixCheck(dupTracker, toBytes(testrecord.input))
		if !cmpDistinctCounts(distinctCounts, testrecord.distinctCounts) || ((len(prefixExample) > 0) != testrecord.prefixExists) {
			// Lazy
			t.Errorf("test number %d failed", num+1)
		}
	}
}
