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
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const VERSION = "0.9b"

// Where the "predefined" wordlists at?
const WORDLIST_DIRECTORY = "wordlists"

// Can't have non-zero entropy if there aren't even two words
const RANDOMNESS_NEEDS_AT_LEAST_TWO_WORDS = 2

// And having 1000 words to be all the same word won't help with entropy either
const RANDOMNESS_NEEDS_DISTINCT_WORDS = 3

// Not all words can be selected to use with dice, but the words are not unique,
// thus selection can't be performed without ambiguity
const UNAMBIGUOUS_TRIM = 216

// Not enough words left to use after trimming the wordlist to the floor that is power of dice sides
const DICE_NOT_USABLE = 219

// The documentation says a error can happen when perusing CRNG, but it doesn't
// say why. It doesn't even say if we can retry
const ERROR_CRNG_TOLD_US_TO_FUCKOFF = 666

// This shall never happen, but stay vigilant
const FATAL_NEGATIVE_ENTROPY_ESTIMATE = 333

func GetFileNameFromDictName(dir string, listname string) string {
	fname := filepath.Join(dir, listname)
	if IsFileExists(fname) {
		return fname
	}
	fname2 := fname + ".asc"
	if IsFileExists(fname2) {
		return fname2
	} else {
		fname2 := fname + ".txt"
		if IsFileExists(fname2) {
			return fname2
		}
		fmt.Printf("There is no such wordlist: \"%s\" installed with the program.\n", listname)
		fmt.Printf("Use 'offend -l' to enumerate wordlists.\n")
		os.Exit(1)
	}
	return "thiscodeisnotreached"
}

func IsFileExists(fname string) bool {
	if st, err := os.Stat(fname); err == nil {
		if st.IsDir() {
			return false
		} else {
			return true
		}
	} else if os.IsNotExist(err) {
		return false
	} else {
		fmt.Printf("Undefined file state when trying to check for file existence: %s.\n", fname)
		os.Exit(1)
	}
	return false // should not be reached
}

func GetReaderForFile(fname string) io.Reader {
	if fname == "-" {
		return os.Stdin
	}
	f, err := os.Open(fname)
	if err != nil {
		fmt.Printf("An error has occured while trying to read %s: %s\n", fname, err)
		os.Exit(1)
	}
	return f
}

func PrintWordLists() {
	dir, err1 := os.Open(WORDLIST_DIRECTORY)
	//ens, err := os.ReadDir(WORDLIST_DIRECTORY) // only since go 1.16
	if err1 != nil {
		fmt.Printf("An error has occured while opening directory %s: %s\n", WORDLIST_DIRECTORY, err1)
		os.Exit(1)
	}
	ens, err2 := dir.Readdir(-1)
	if err2 != nil {
		fmt.Printf("An error has occured while trying to list files in directory %s: %s\n", WORDLIST_DIRECTORY, err2)
		os.Exit(1)
	}
	wl := make([]string, 0, 0)
	for _, ensie := range ens {
		if !ensie.IsDir() {
			fname := ensie.Name()
			if strings.HasSuffix(fname, ".asc") || strings.HasSuffix(fname, ".txt") {
				// can use string length, cause the last characters are one-byte
				fname = fname[:len(fname)-4]
			}
			wl = append(wl, fname)
		}
	}
	sort.Sort(sort.StringSlice(wl))
	for i := 0; i < len(wl); i++ {
		if i == 0 || wl[i] != wl[i-1] {
			fmt.Println(wl[i])
		}
	}
}

func complainAboutTrimAndExit(totalWords int, usableWords int) {
	fmt.Printf("The %d is not a power of %d. The floor of %d that is a power of %d is %d.\n", totalWords, sysConfig.DiceFaces, totalWords, sysConfig.DiceFaces, usableWords)
	fmt.Printf("However, some words are occuring twice or more, thus the selection of %d words out of %d words can't be performed unambiguously.\n", usableWords, totalWords)
	os.Exit(UNAMBIGUOUS_TRIM)
}

func main() {
	configure()
	if sysConfig.Verbosity > 0 {
		fmt.Printf("Offend ver %s (c) VigilantDoomer, 2023. All rights reserved.\n", VERSION)
	}
	if sysConfig.ListWordLists {
		PrintWordLists()
		os.Exit(0)
	}

	preamble := false

	// configure user-chosen random passphrase generator with user-chosen delimiter
	// and number of dice faces (if applicable)
	currentRnd := NewRndSource(sysConfig.RndSource)
	currentRnd.SetDelimiter(sysConfig.Delimiter)
	switch c := currentRnd.(type) {
	case RndSourceWithDice:
		c.SetDiceFaces(sysConfig.DiceFaces)
		preamble = true
	}

	// dictionary file, depending on the option used, is either identified directly by filename
	// or is identified by a "dictionary name", which is a name of a file (possibly omitting its extension)
	// in a directory relative to the running program, this directory's name being hardcoded constant
	fname := sysConfig.DictFileName
	if fname == "" {
		fname = GetFileNameFromDictName(WORDLIST_DIRECTORY, sysConfig.WordListName)
	}
	rd := GetReaderForFile(fname)

	dupTracker := make(map[string]int)
	words, gotUpperCaseLettersInSource, wordLenTotal := parseWords(rd, dupTracker)
	uniqueWords := len(dupTracker)
	// How many unique word frequencies are there, and whether any words
	// are prefixes of any other
	allCnts, prefixData := getDistinctCountsAndDoPrefixCheck(dupTracker, words)

	// Abort early if not enough words
	if len(words) < 2 {
		fmt.Printf("Discovered only %d words.\n", len(words))
		fmt.Println("Can't generate random output with less than 2 words - exiting.")
		os.Exit(RANDOMNESS_NEEDS_AT_LEAST_TWO_WORDS)
	}

	// At least two words need to be distinct
	if uniqueWords == 1 {
		fmt.Println("All words are the same word.")
		fmt.Println("For output to be random, at least 2 words must be distinct - exiting.")
		os.Exit(RANDOMNESS_NEEDS_DISTINCT_WORDS)
	}

	// Entropy estimation.
	var entropyPerWord float64 = 0.0
	usableWordsNum := currentRnd.Usable(len(words))
	if usableWordsNum <= 1 {
		fmt.Println("This number of dice sides can't be used with this dictionary.")
		os.Exit(DICE_NOT_USABLE)
	}

	// TODD refactor / cover entire entropy logic with tests.
	if len(allCnts) == 1 {
		if usableWordsNum == len(words) {
			entropyPerWord = math.Log2(float64(uniqueWords))
		} else if len(words) != uniqueWords {
			complainAboutTrimAndExit(len(words), usableWordsNum)
		}
	} else {
		if usableWordsNum != len(words) {
			complainAboutTrimAndExit(len(words), usableWordsNum)
		}
		for i := 0; i < len(allCnts); i++ {
			thisWordFreq := float64(allCnts[i][0]) / float64(len(words))
			numOfWordsForThisFreq := allCnts[i][1]
			entropyPerThisWord := float64(-1.0) * math.Log2(thisWordFreq) * thisWordFreq * float64(numOfWordsForThisFreq)
			entropyPerWord = entropyPerWord + entropyPerThisWord
		}
	}

	entropyTarget := sysConfig.Entropy
	numWordsToGenerate := sysConfig.NumWords
	if numWordsToGenerate == 0 {
		numWordsFraq := entropyTarget / entropyPerWord
		if entropyPerWord < 0 {
			fmt.Printf("Fatal error: got negative entropy per word %f\n.", entropyPerWord)
			os.Exit(FATAL_NEGATIVE_ENTROPY_ESTIMATE)
		}

		numWordsToGenerate = int64(math.Ceil(numWordsFraq))
		if float64(numWordsToGenerate)*entropyPerWord < entropyTarget {
			// I may be dumb, but I want to be sure
			numWordsToGenerate++
		}
	}

	if sysConfig.Verbosity > 0 {
		fmt.Printf("Read in %d words. Of them %d are unique.\n", len(words), uniqueWords)
	}

	if usableWordsNum != len(words) {
		fmt.Printf("Using only %d words out of %d. This happened because the number of words in the dictionary is not a power of dice sides number (%d).\n", usableWordsNum, len(words), sysConfig.DiceFaces)
		if sysConfig.NumWords != 0 {
			fmt.Println("Given that you specified the number of words to generate directly, you are getting reduced entropy compared to using full list.")
		}
	}

	if sysConfig.Verbosity > 0 {
		if len(allCnts) == 1 {
			fmt.Println("Word distribution is fair (good).")
		} else {
			fmt.Println("Word distribution is NOT fair.")
			fmt.Println("Some words occur more frequently than the others.")
		}
		// Valid only if uniquely decodeable
		fmt.Printf("Entropy per word: %f\n", entropyPerWord)
	}

	// Hide average word length and average entropy per character behind
	// verbose flag. These parameters are a measure of AVERAGE convenience
	// the dictionary provides, and may not correlate with a particular
	// result user gets, so avoiding them decreases the amount of confusion
	// they can cause
	// Careful with measuring unicode string length (this is handled elsewhere)
	if sysConfig.Verbosity >= 2 {
		avgWordLen := float64(wordLenTotal) / float64(len(words))
		avgCharEntropy := entropyPerWord / avgWordLen
		fmt.Printf("Average word length: %f\n", avgWordLen)
		fmt.Printf("Average entropy per character: %f\n", avgCharEntropy)
	}

	if gotUpperCaseLettersInSource {
		fmt.Println("Hint: Some dictionary words contain uppercase letters.")
		preamble = true
	} else {
		if sysConfig.Verbosity > 0 {
			fmt.Println("No dictionary words contain uppercase letters (good).")
		}
	}

	if prefixData == nil {
		if sysConfig.Verbosity > 0 {
			fmt.Println("No words are prefixes of others (good).")
		}
	} else {
		preamble = true
		// Use SardinasPatterson to check if the dictionary can still generate secure enough passphrases,
		// if used with caution
		fmt.Println("The list includes some words that are prefixes of others.")
		fmt.Printf("Example: word \"%s\" is a prefix of word \"%s\".\n", prefixData[0][0], prefixData[0][1])
		fmt.Printf("Checking for whether the passphrases are all uniquely decodeable nonetheless... (this might take some time)\n")
		if SardinasPatterson_IsSafe(words) {
			fmt.Println("YES, passphrases are all uniquely decodeable. (GOOD)")
			fmt.Println("Warning: You need to type the generated passphrase verbatim, otherwise unique decodability might CEASE to hold.")
		} else {
			// Neither picking a delimiter nor changing words' case guarantees a solution. Delimiter can be present in some words,
			// words can have different casing that result in new collisions after conversions.
			fmt.Println("NO, passphrases are not uniquely decodable. (BAD)")
			fmt.Println("Warning: The enthropy estimate is invalid, the security of your passphrase is LOWER than requested.")
		}
	}
	preamble = preamble || (sysConfig.Verbosity > 0)
	if preamble {
		fmt.Printf("Will generate %d words.\n", numWordsToGenerate)
	}
	fmt.Println(currentRnd.Generate(words, numWordsToGenerate))
}
