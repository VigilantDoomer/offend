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

// TODO this would love also:
// 1. Fuzz testing
// 2. Benchmarks
//
// When coming up with tests, cue that:
// 1. Any dictionary that is prefix code is also safe according to Sardinas-Patterson algorithm (SardinasPatterson_IsSafe must return true)
// 2. Any dictionary that is NOT safe according to Sardinas-Patterson algorithm (SardinasPatterson_IsSafe return false), is NOT prefix code
// 3. Some dictionaries that are NOT prefix code are nonetheless safe according to Sardinas-Patterson algorithm, exactly the reason I've implemented it
import (
	"testing"
)

type sardinasPatterson_testrecord struct {
	input  []string
	result bool
}

func TestSardinasPatterson_IsSafe(t *testing.T) {
	dataset := []sardinasPatterson_testrecord{
		sardinasPatterson_testrecord{input: []string{"dog", "cat", "doggerel"}, result: true},
		sardinasPatterson_testrecord{input: []string{"dog", "dog", "doggerel", "dog", "doggerel"}, result: true},
		sardinasPatterson_testrecord{input: []string{"dog", "cat", "dogcat"}, result: false},
		sardinasPatterson_testrecord{input: []string{"dog", "cat", "donut", "nutjob", "nutcat", "catdoggery"}, result: true},
		sardinasPatterson_testrecord{input: []string{"dog", "cat", "doggerel", "gerelcat"}, result: false},
		sardinasPatterson_testrecord{input: []string{"doggie", "catmonster", "doggerel", "doggy"}, result: true},
	}
	for num, testrecord := range dataset {
		res := SardinasPatterson_IsSafe(toBytes(testrecord.input))
		if res != testrecord.result {
			t.Errorf("test number %d failed\n   got: %t\n   expected: %t\n", num+1, res, testrecord.result)
		}
	}
}
