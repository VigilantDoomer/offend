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
	"os"

	"github.com/spf13/pflag"
)

type RandomSource int

const (
	CryptoPRNG RandomSource = iota
	RealDice
)

type Config struct {
	NumWords      int64
	Verbosity     int
	Entropy       float64
	Delimiter     string
	Capitalize    bool
	WordListName  string
	ListWordLists bool
	DictFileName  string
	DiceFaces     int
	RndSource     RandomSource
}

var sysConfig *Config = nil

// what a mess
func checkForMutualExclusiveFlags() {
	mutexed := []string{"en", "wl"}
	visited := ""
	shortToLong := map[string]string{}
	pflag.Visit(func(flg *pflag.Flag) {
		visited = visited + flg.Shorthand
		shortToLong[flg.Shorthand] = flg.Name
	})
	for _, s := range mutexed {
		a := 0
		prevch := ""
		for _, ch := range visited {
			for _, ch2 := range s {
				if ch == ch2 {
					a++
					if a == 2 {
						fmt.Printf("Parameters -%s (--%s) and -%s (--%s) are mutually exclusive.\n", prevch, shortToLong[prevch], string(ch), shortToLong[string(ch)])
						os.Exit(1)
					} else {
						prevch = string(ch)
					}
				}
			}
		}
	}
}

func configure() {
	con := new(Config)
	strRndSource := ""
	pflag.Int64VarP(&(con.NumWords), "num", "n", 0, "Number of words to concatenate.")
	pflag.Float64VarP(&(con.Entropy), "entropy", "e", 77.5, "Desired entropy, in bits.")
	pflag.StringVarP(&(con.Delimiter), "delimiter", "d", "", "Separate words by delimiter. Empty string by default")
	pflag.BoolVarP(&(con.Capitalize), "caps", "c", true, "Capitalize words.")
	pflag.StringVarP(&(con.WordListName), "wordlist", "w", "offend_fast", "Use words from this wordlist.")
	pflag.BoolVarP(&(con.ListWordLists), "list", "l", false, "List all the available wordlists which can be passed to -w (--wordlist) parameter")
	pflag.StringVarP(&(strRndSource), "randomsource", "r", "system", "Get randomness from this source. Possible values: \"realdice\", \"system\".")
	pflag.IntVarP(&(con.DiceFaces), "faces", "f", 6, "Number of faces/sides of dice, when \"realdice\" is used as source.")
	pflag.CountVarP(&(con.Verbosity), "verbose", "v", "Be verbose. Use several times for increased verbosity.")
	pflag.Parse()
	ss := pflag.Args()
	if len(ss) > 0 {
		con.DictFileName = ss[0]
	} else {
		con.DictFileName = ""
	}
	if strRndSource == "system" {
		con.RndSource = CryptoPRNG
	} else if strRndSource == "realdice" {
		con.RndSource = RealDice
	} else if strRndSource != "" {
		fmt.Printf("Unknown random source: '%s'. Should be 'realdice' or 'system' (case-sensitive)\n", strRndSource)
		os.Exit(1)
	}
	checkForMutualExclusiveFlags()
	sysConfig = con
}
