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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// A regular expression with one capturing subgroup that should match a word in either
// a numbered dictionary entry or an unnumbered one.
var FIND_WORD_REGEX *regexp.Regexp = regexp.MustCompile(`^[0-9]+(?:-[0-9]+)*\s+([^\s]+)$`)

// Parses input dictionary, stores and indexes all words into
// a slice for fast lookup
func parseWords(rd io.Reader, dupTracker map[string]int) ([][]byte, bool, int) {
	// Arbitrary slice initial size - fix later
	ret := make([][]byte, 0, 7770)
	sc := bufio.NewScanner(rd)
	firstLineChecked := false
	signed := false
	skipUntilEmptyLine := false
	lineNum := 0
	wordLenTotal := 0
	gotUpperCaseLettersInSource := false
	for sc.Scan() {
		// Is scanner in a good shape or not
		err := sc.Err()
		if err != nil {
			// TODO refer to input by name, possibly on a separate line so as not
			// to obscure line number and error message. Somewhere between "input line"
			// and "scan aborted" message
			fmt.Printf("Scan: input line %d: encountered an error: %s.\n", lineNum, err)
			fmt.Println("Dictionary scan aborted, discarding remaining words.")
			break
		}

		// Is this a PGP signed dictionary?
		if !firstLineChecked {
			firstLineChecked = true
			rTrimmed := bytes.TrimRightFunc(sc.Bytes(), unicode.IsSpace)
			signed = bytes.Equal(rTrimmed, []byte("-----BEGIN PGP SIGNED MESSAGE-----"))
			if signed {
				skipUntilEmptyLine = true
				continue
			}
		}

		// Beware: the underlying array for "line" slice lies in buffer
		// allocated by *bufio.Scanner. It may be overwritten by any of
		// the subsequent calls to Scan()
		line := sc.Bytes()

		// For PGP signed dictionary, we need to skip the header, whose end
		// is marked by empty line
		if skipUntilEmptyLine {
			if line != nil && len(line) > 0 {
				continue
			} else {
				skipUntilEmptyLine = false
				// Done (skipping PGP message at the beginning of the file).
			}
		}

		// parseOneWord allocates new memory for result
		wrd := parseOneWord(line, signed)
		if wrd != nil {
			// For PGP signed dictionary, we stop processing upon encountering
			// PGP signature
			if signed && bytes.Equal(wrd, []byte("-----BEGIN PGP SIGNATURE-----")) {
				break
			}
			gotUpperCaseLettersInSource = gotUpperCaseLettersInSource || hasUpperCaseChars(wrd)
			// If capitalization of first letters is performed, the words MUST
			// be compared in capitalized form throughout the program, because if the dictionary
			// contains both words that begin with a capital letter and those that don't,
			// capitalization may introduce DUPLICATES and PREFIX PROBLEM where there
			// were NONE.
			if sysConfig.Capitalize {
				patchFirstLetterToUpperCase_InPlace(wrd)
			}
			// Add this word to result list
			ret = append(ret, wrd)
			// Increment the number this word has been seen, if any
			tmpShit := string(wrd)
			wordLenTotal = wordLenTotal + utf8.RuneCountInString(tmpShit)
			oldCnt, _ := dupTracker[tmpShit]
			dupTracker[tmpShit] = oldCnt + 1
		}
		lineNum = lineNum + 1
	}
	return ret, gotUpperCaseLettersInSource, wordLenTotal
}

// Returns true if two byte slices reference same place in memory
// Don't pass empy slices or will trigger exception
func sameRef_NEBS(nonEmptyByteSlice1, nonEmptyByteSlice2 []byte) bool {
	return &nonEmptyByteSlice1[0] == &nonEmptyByteSlice2[0]
}

// Finds a word within a line of dictionary
// If a line does not contain a word, returns nil
// Otherwise returns the found word
// MUST allocate new bytes for non-nil result,
// because wrd slice refers to memory inside a Scanner's buffer
func parseOneWord(wrd []byte, signed bool) []byte {
	if wrd == nil {
		return nil
	}
	if signed && bytes.HasPrefix(wrd, []byte("- ")) {
		wrd = wrd[2:]
	}
	// Remove whitespace left and right
	wrd = bytes.TrimSpace(wrd)
	if wrd == nil || len(wrd) == 0 {
		return nil
	}
	smatch := FIND_WORD_REGEX.FindSubmatch(wrd)
	// It's called volatile, because it refers
	// to the current token in Scanner's buffer
	var volatile_data []byte
	if smatch == nil {
		// Couldn't parse out a word, will assume
		// the whole line is one dictionary "word"
		volatile_data = wrd
	} else {
		volatile_data = smatch[1]
	}

	// Perform normalization under NFC - affects
	// 1. which dictionary words will be considered equal,
	// 2. recognition of words that are prefixes of
	// other words
	// Reference: Unicode® Standard Annex #15
	// https://unicode.org/reports/tr15/
	// NOTE. This measure, by itself, is not necessary
	// sufficient to achieve correctness for both
	normalizedData := norm.NFC.Bytes(volatile_data)

	// Must return slice that does not share memory
	// with wrd (parseWords got it pointing inside
	// Scanner's buffer before passing to us)
	var ret []byte
	if sameRef_NEBS(normalizedData, volatile_data) {
		ret = make([]byte, len(volatile_data))
		copy(ret, volatile_data)

	} else {
		// New bytes were already allocated to hold
		// normalized data, they are safe to return
		ret = normalizedData
	}
	return ret
}

// Overwrites the first letter of the byte slice so it is uppercase
// Doesn't return anything
// Auxilary memory allocation may actually happen, but the
// original slice points to the same memory address, while
// the contents are altered
// ? avoid memory allocation maybe
func patchFirstLetterToUpperCase_InPlace(word []byte) {
	r, firstRuneSize := utf8.DecodeRune(word)
	if r != utf8.RuneError {
		sl2 := bytes.ToUpper(word[0:firstRuneSize])
		if !sameRef_NEBS(sl2, word) {
			copy(word[0:firstRuneSize], sl2[0:firstRuneSize])
		}
	}
}
