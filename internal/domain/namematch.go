package domain

import (
	"sort"
	"strings"
	"unicode"
)

// nameFolder maps German umlauts to their common transliteration and strips
// frequent accents, so "Müller", "Mueller" and "Muller" compare closely.
var nameFolder = strings.NewReplacer(
	"ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss",
	"á", "a", "à", "a", "â", "a", "å", "a", "ã", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ø", "o",
	"ú", "u", "ù", "u", "û", "u",
	"ç", "c", "ñ", "n", "ł", "l", "š", "s", "ž", "z", "č", "c", "ć", "c",
)

// nameTokens lower-cases, folds and splits a name into its parts, dropping
// punctuation ("Müller-Lüdenscheidt, Hans" → [mueller luedenscheidt hans]).
func nameTokens(name string) []string {
	folded := nameFolder.Replace(strings.ToLower(name))
	return strings.FieldsFunc(folded, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// NameSimilarity rates how alike two person names are, from 0 (nothing in
// common) to 1 (same name), for suggesting members to a free-text kiosk entry
// (K-015). It tolerates typos, umlaut spellings, swapped first and last
// names and shortened first names ("Max" ↔ "Maximilian").
func NameSimilarity(a, b string) float64 {
	ta, tb := nameTokens(a), nameTokens(b)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}
	best := stringSimilarity(strings.Join(ta, " "), strings.Join(tb, " "))
	if s := stringSimilarity(sortedJoin(ta), sortedJoin(tb)); s > best {
		best = s
	}
	if s := tokenSimilarity(ta, tb); s > best {
		best = s
	}
	return best
}

func sortedJoin(tokens []string) string {
	cp := append([]string(nil), tokens...)
	sort.Strings(cp)
	return strings.Join(cp, " ")
}

// tokenSimilarity pairs every part of the shorter name with its best match in
// the longer one (a prefix counts as a strong match) and averages the scores,
// lightly penalising parts of the longer name left unmatched.
func tokenSimilarity(a, b []string) float64 {
	if len(a) > len(b) {
		a, b = b, a
	}
	used := make([]bool, len(b))
	var sum float64
	for _, x := range a {
		bestScore, bestIdx := 0.0, -1
		for j, y := range b {
			if used[j] {
				continue
			}
			s := stringSimilarity(x, y)
			if len(x) >= 2 && len(y) >= 2 && (strings.HasPrefix(y, x) || strings.HasPrefix(x, y)) && s < 0.9 {
				s = 0.9
			}
			if s > bestScore {
				bestScore, bestIdx = s, j
			}
		}
		if bestIdx >= 0 {
			used[bestIdx] = true
		}
		sum += bestScore
	}
	score := sum / float64(len(a))
	// Each part of the longer name without a counterpart costs a quarter: a
	// bare first name is a lead, not a confident match.
	return max(0, score*(1-0.25*float64(len(b)-len(a))))
}

// stringSimilarity is 1 - normalised Levenshtein distance.
func stringSimilarity(a, b string) float64 {
	ra, rb := []rune(a), []rune(b)
	maxLen := max(len(ra), len(rb))
	if maxLen == 0 {
		return 1
	}
	return 1 - float64(levenshtein(ra, rb))/float64(maxLen)
}

func levenshtein(a, b []rune) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}
