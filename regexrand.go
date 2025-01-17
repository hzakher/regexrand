package regexrand

import (
	"crypto/rand"
	"math/big"
	"regexp/syntax"
	"strconv"
	"strings"
	"unicode"
)

func cryptoRandInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Uint64())
}

func asciiExcluding(xchars string) string {
	var res strings.Builder
	// Only use ASCII 32 to 126
	for i := rune(32); i < 127; i++ {
		if strings.ContainsRune(xchars, i) {
			res.WriteRune(i)
		}
	}
	return res.String()
}

// GenerateMatch writes to string builder b, whose result matches a given
// regular expression re. Using "or more" operators will have a given limit
// moreLimit.
func GenerateMatch(b *strings.Builder, re *syntax.Regexp, moreLimit int) {
	switch re.Op {
	default:
		b.WriteString("<invalid op" + strconv.Itoa(int(re.Op)) + ">")
	case syntax.OpNoMatch:
		b.WriteString(`<No Match operator is unsupported>`)
	case syntax.OpEmptyMatch:
		b.WriteString("")
	case syntax.OpLiteral:
		for _, r := range re.Rune {
			b.WriteRune(r)
		}
	case syntax.OpCharClass:
		if len(re.Rune)%2 != 0 {
			b.WriteString(`[invalid char class]`)
			break
		} else if re.Rune[0] == 0 && re.Rune[len(re.Rune)-1] == unicode.MaxRune {
			// None of char
			var charset strings.Builder
			for i := 0; i < len(re.Rune); i += 2 {
				lo, hi := re.Rune[i], re.Rune[i+1]
				if lo != hi {
					for ; lo < hi+1; lo++ {
						charset.WriteRune(lo)
					}
				} else {
					charset.WriteRune(lo)
				}
			}
			stringset := asciiExcluding(charset.String())
			charbyte := stringset[cryptoRandInt(len(stringset)-1)]
			b.WriteByte(charbyte)
		} else {
			// Any of char
			var charset strings.Builder
			for i := 0; i < len(re.Rune); i += 2 {
				lo, hi := re.Rune[i], re.Rune[i+1]
				if lo != hi {
					for ; lo < hi+1; lo++ {
						charset.WriteRune(lo)
					}
				} else {
					charset.WriteRune(lo)
				}
			}
			stringset := charset.String()
			charbyte := stringset[cryptoRandInt(len(stringset)-1)]
			b.WriteByte(charbyte)
		}
	case syntax.OpAnyCharNotNL, syntax.OpAnyChar:
		// Only use 32 to 126 ASCII so no NL anyways.
		stringset := asciiExcluding("")
		charbyte := stringset[cryptoRandInt(len(stringset)-1)]
		b.WriteByte(charbyte)
	case syntax.OpBeginLine:
		if b.Len() > 0 { // If builder is not already just NL
			b.WriteByte('\n')
		}
	case syntax.OpEndLine:
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
	case syntax.OpBeginText:
		if b.Len() > 0 {
			b.WriteString(`<Invalid position of OpBeginText. Needs to be at the start of expression>`)
		}
	case syntax.OpEndText:
		break
	case syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		b.WriteString(`<Word boundary operators are unsupported>`)
	case syntax.OpCapture:
		// Generate capture body. Capture groups can be found later with
		// FindStringSubmatch() and SuxexpNames() on result.
		if re.Sub[0].Op != syntax.OpEmptyMatch {
			GenerateMatch(b, re.Sub[0], moreLimit)
		}
	case syntax.OpStar, syntax.OpPlus:
		min := 0
		if re.Op == syntax.OpPlus {
			min = 1
		}
		for i := min + cryptoRandInt(moreLimit-min+1); i > 0; i-- {
			GenerateMatch(b, re.Sub[0], moreLimit)
		}
	case syntax.OpQuest:
		if cryptoRandInt(0x7FFFFFFF) > 0x3FFFFFFF {
			GenerateMatch(b, re.Sub[0], moreLimit)
		}
	case syntax.OpRepeat:
		for i := re.Min + cryptoRandInt(re.Max-re.Min+1); i > 0; i-- {
			GenerateMatch(b, re.Sub[0], moreLimit)
		}
	case syntax.OpConcat:
		for _, sub := range re.Sub {
			GenerateMatch(b, sub, moreLimit)
		}
	case syntax.OpAlternate:
		GenerateMatch(b, re.Sub[cryptoRandInt(len(re.Sub)-1)], moreLimit)
	}
}

// GenerateString helper function to generate a match and return a string
func GenerateString(re *syntax.Regexp, moreLimit int) string {
	var b *strings.Builder
	GenerateMatch(b, re, moreLimit)
	item := b.String()
	b.Reset()
	return item
}

// GenerateStrings helper function to generate a set of matchs (count) and return a string array
// generated set may have duplicate values
// for unique set, use GenerateUniqueStrings
func GenerateStrings(re *syntax.Regexp, moreLimit, count int) []string {
	result := make([]string,0)
	i := 0
	for i < count {	
		item := GenerateString(re, moreLimit)
		result = append(result, item)
	}
	return result
}

// GenerateUniqueStrings helper function to generate a set of matchs (count) and return a string array.
// generated set will have unique values
func GenerateUniqueStrings(re *syntax.Regexp, moreLimit, count int) []string {
	
	cache := make(map[string]int)
	result := make([]string,0)
	i, ii := 0,0
	iMaxRun := 10000000 
	for i < count {
		ii++
		// if ii reached the maximum run, then break the loop and return the result.
		// this is to avoid infint loop where the count of unique values cannot reach "matchCount"
		if ii >= iMaxRun {
			// maximum run reached, return the items we already identified
			return result
		}
		
		item := GenerateString(re, moreLimit)
		//check if item is in the result.
		if _, ok := cache[item];!ok {
		  result = append(result, item)		
		  cache[item]=1
		  i++
		}
	}
	return result
}
