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
	"io"
	"testing"
)

type parseOneWord_testrecord struct {
	wrd    string
	signed bool
	result string
}

type parseWords_testrecord struct {
	input            []string
	configCapitalize bool
	words            []string
	hasCaps          bool
}

func TestParseOneWord(t *testing.T) {
	dataset := []parseOneWord_testrecord{
		parseOneWord_testrecord{wrd: "111 vigilant", signed: false, result: "vigilant"},
		parseOneWord_testrecord{wrd: "solstice", signed: false, result: "solstice"},
		parseOneWord_testrecord{wrd: "- mortem", signed: true, result: "mortem"},
		parseOneWord_testrecord{wrd: "- 111 jupiter", signed: true, result: "jupiter"},
		parseOneWord_testrecord{wrd: "111", signed: false, result: "111"},
	}
	for _, testrecord := range dataset {
		res := parseOneWord([]byte(testrecord.wrd), testrecord.signed)
		if string(res) != testrecord.result {
			t.Errorf("parsed record: \n   line: %s\n   signed: %t\n   got: %s\n   expected: %s\n", testrecord.wrd, testrecord.signed, res, testrecord.result)
		}
	}
}

// joins input strings placing newlines inbetween,
// return reader from the string obtained by such join
func stringArrayToTextIo(input []string) io.Reader {
	buf := new(bytes.Buffer)
	for i, s := range input {
		buf.WriteString(s)
		if i < len(input)-1 {
			buf.WriteByte(byte('\n'))
		}
	}
	return bytes.NewReader(buf.Bytes())
}

func cmpStringsToBytes(strs []string, bts [][]byte) bool {
	if len(strs) != len(bts) {
		return false
	}
	for i, v := range strs {
		if v != string(bts[i]) {
			return false
		}
	}
	return true
}

func cmpParseWordsResult(testrecord parseWords_testrecord, words [][]byte, hasCaps bool) bool {
	return (hasCaps == testrecord.hasCaps) && cmpStringsToBytes(testrecord.words, words)
}

func TestParseWords(t *testing.T) {
	// provide a mostly bogus config so things don't break on trying
	// to access it
	defConfig := &Config{NumWords: 0,
		Verbosity:     0,
		Entropy:       0,
		Delimiter:     "",
		Capitalize:    false,
		WordListName:  "",
		ListWordLists: false,
		DictFileName:  "",
		DiceFaces:     6,
		RndSource:     CryptoPRNG}
	// set global config to it
	sysConfig = defConfig
	dataset := []parseWords_testrecord{
		parseWords_testrecord{input: []string{"111 vigilant", "112 solstice", "113 mortem", "114 jupiter"},
			configCapitalize: true,
			words:            []string{"Vigilant", "Solstice", "Mortem", "Jupiter"},
			hasCaps:          false},
		parseWords_testrecord{input: []string{"-----BEGIN PGP SIGNED MESSAGE-----",
			"Hash: oops (not inventing for the test)", "", "111 vigilant", "112 solstice", "113 mortem", "114 jupiter", "115 ashore", "-----BEGIN PGP SIGNATURE-----",
			"", "Lol! just kidding. No sig here", "-----END PGP SIGNATURE-----"},
			configCapitalize: false,
			words:            []string{"vigilant", "solstice", "mortem", "jupiter", "ashore"},
			hasCaps:          false},
	}
	for num, testrecord := range dataset {
		sysConfig.Capitalize = testrecord.configCapitalize
		dupTracker := make(map[string]int)
		inputText := stringArrayToTextIo(testrecord.input)
		words, hasCaps, _ := parseWords(inputText, dupTracker)
		if !cmpParseWordsResult(testrecord, words, hasCaps) {
			// Lazy, but then again, these are supposed to be whole texts
			// Oh yeah, and human numbers start from 1, unlike machine numbers
			t.Errorf("Failure in test #%d\n", num+1)
		}
	}
}
