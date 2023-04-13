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
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
)

type RndSource interface {
	SetDelimiter(d string)
	Usable(totalWords int) int
	Generate(words [][]byte, numWordsToGenerate int64) string
}

type RndSourceWithDice interface {
	RndSource
	SetDiceFaces(faces int)
}

type CryptoPRNGImpl struct {
	delim string
}

type RealDiceImpl struct {
	delim string
	faces int
}

func NewRndSource(rndSource RandomSource) RndSource {
	if rndSource == CryptoPRNG {
		return new(CryptoPRNGImpl)
	} else if rndSource == RealDice {
		return new(RealDiceImpl)
	}
	fmt.Printf("Program error: unknown rndSource %d.\n", int(rndSource))
	return nil
}

func (c *CryptoPRNGImpl) SetDelimiter(d string) {
	c.delim = d
}

func (c *CryptoPRNGImpl) Usable(totalWords int) int {
	return totalWords
}

func (c *CryptoPRNGImpl) Generate(words [][]byte, numWordsToGenerate int64) string {
	passBuilder := strings.Builder{}
	for i := int64(0); i < numWordsToGenerate; i++ {
		chRand, err := cRand_UInt(uint(len(words)))
		if err != nil {
			fmt.Printf("Cryptographic pseudo random generation failed: %s.\n", err.Error())
			os.Exit(ERROR_CRNG_TOLD_US_TO_FUCKOFF)
		}
		cho_word := words[chRand]
		passBuilder.Write(cho_word)
		if c.delim != "" && i < (numWordsToGenerate-1) {
			passBuilder.WriteString(c.delim)
		}
	}
	return passBuilder.String()
}

func (r *RealDiceImpl) SetDelimiter(d string) {
	r.delim = d
}

func (r *RealDiceImpl) SetDiceFaces(faces int) {
	r.faces = faces
}

func (r *RealDiceImpl) getDicePerWord(totalWords int) float64 {
	fracDicePerWord := math.Log2(float64(totalWords)) / math.Log2(float64(r.faces))
	return math.Floor(fracDicePerWord)
}

func (r *RealDiceImpl) Usable(totalWords int) int {
	dpw := r.getDicePerWord(totalWords)
	return int(math.Pow(float64(r.faces), dpw))
}

// needed to write my wrapped, because Scanln with a *int
// behaves stupid
func readInt(greeting string) (int, error) {
	fmt.Printf(greeting)
	val := ""
	var err error
	var ret int
	_, err = fmt.Scanln(&val)
	if err != nil {
		return 0, err
	}
	ret, err = strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (r *RealDiceImpl) chooseWord(words [][]byte, limit int, dicePerWord int) int {
	ret := 0
	for {
		ret = 0
		for i := 0; i < dicePerWord; i++ {
			rolled := 0
			for (rolled < 1) || (rolled > r.faces) {
				var err error
				rolled, err = readInt(fmt.Sprintf("What number shows dice number %d? ", i+1))
				if err != nil {
					fmt.Printf("The value was not valid. %s\n", err.Error())
					rolled = 0
				} else if (rolled < 1) || (rolled > r.faces) {
					fmt.Printf("Value out of range: %d, should be >=%d and <=%d\n", rolled, 1, r.faces)
					rolled = 0
				}
				ret = ret + int(math.Pow(float64(r.faces), float64(i-1)))*rolled
			}
		}
		if ret < limit {
			break
		} else {
			fmt.Printf("Value out of range. Please roll dice again.")
		}
	}
	return ret
}

func (r *RealDiceImpl) Generate(words [][]byte, numWordsToGenerate int64) string {
	passBuilder := strings.Builder{}
	totalWords := len(words)
	dpw := int(r.getDicePerWord(totalWords))
	usableWordsNum := r.Usable(totalWords)
	for i := int64(0); i < numWordsToGenerate; i++ {
		fmt.Printf("Generating word number %d:\n", i+1)
		cho_word := words[r.chooseWord(words, usableWordsNum, dpw)]
		passBuilder.Write(cho_word)
		if r.delim != "" && i < (numWordsToGenerate-1) {
			passBuilder.WriteString(r.delim)
		}
	}
	return passBuilder.String()
}

// Wrapper over crypto.rand.Int that uses
// uint type instead of big.NewInt
func cRand_UInt(num uint) (uint, error) {
	numBig := big.NewInt(int64(num))
	chRand, err := rand.Int(rand.Reader, numBig)
	if err != nil {
		return 0, err
	}
	return uint(chRand.Uint64()), nil
}
