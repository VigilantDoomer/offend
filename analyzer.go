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
	"bytes"
	"sort"
	"strings"
	"unicode"
)

// Produce a slice with all distinct count of words in the dictionary
// ... it was tempting to omit "o" in "counts"
func getDistinctCountsAndDoPrefixCheck(dupTracker map[string]int, words [][]byte) ([][]int, [][]string) {
	cnts := make([][]int, 0)
	srt := make([]string, 0)
	for k, v := range dupTracker {
		cnts = addIfUnique_ASC(cnts, v)
		srt = append(srt, k)
	}
	sort.StringSlice(srt).Sort()
	prefixData := doPrefixCheck(srt)
	return cnts, prefixData
}

// Add int into []int, but only if it is unique
// Assume []int is sorted from min to max (ASCending order)
// and keep it that way
func addIfUnique_ASC(sl [][]int, v int) [][]int {
	if len(sl) == 0 {
		// Bruh, I'm alone here yet
		return append(sl, []int{v, 1})
	}
	// Skim over all the inferior numbers
	ii := 0
	for i := 0; i < len(sl) && v > sl[i][0]; i++ {
		ii = i
	}
	// Am I the biggest daddy here?
	if ii < len(sl) {
		// Looks like I'm not...
		// Is it a match?
		if v == sl[ii][0] {
			// Hell yeah... but that means I'm dismissed
			// - not before being counted, though
			sl[ii][1] = sl[ii][1] + 1
			return sl
		} else {
			// make sure slice has enough capacity
			// by appending bogus entry
			sl = append(sl, []int{0, 0})
			// shift all elements to the right by 1
			copy(sl[ii+1:], sl[ii:])
			// get into appropriate pose... er, position
			sl[ii] = []int{v, 1}
		}
	} else {
		// My cock is the largest, it belongs at the end
		// of ascending queue
		sl = append(sl, []int{v, 1})
	}
	return sl
}

func doPrefixCheck(srt []string) [][]string {
	for i := 0; i < len(srt)-1; i++ {
		if strings.HasPrefix(srt[i+1], srt[i]) {
			return [][]string{{srt[i], srt[i+1]}}
		}
	}
	return nil
}

func hasUpperCaseChars(st []byte) bool {
	for _, v := range bytes.Runes(st) {
		if unicode.IsUpper(v) {
			return true
		}
	}
	return false
}
