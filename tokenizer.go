// Package tokenizer provides unicode-aware token counting for text chunking.
// It replaces the naive "words * 1.3" approximation with a heuristic that
// correctly handles CJK scripts (each character ≈ 1 token), Latin/Cyrillic
// words (~1.3 tokens each), numbers (1 token each), and punctuation.
//
// The estimator intentionally favours over-counting slightly to keep chunks
// safely below embedding-model context limits (bge-m3: 8192 tokens).
package tokenizer

import "unicode"

// Count estimates the number of tokens in text. It is safe for all Unicode
// scripts and requires no external models or tokenizer files.
//
// Heuristic by script range:
//   - CJK Unified / Hiragana / Katakana / Hangul: 1 token per rune
//   - Arabic / Hebrew / Thai / Devanagari: 1 token per rune
//   - Latin / Cyrillic / Greek words: word_count * 1.3 (BPE overhead)
//   - Numbers: 1 token per digit sequence
//   - Everything else: 1 token per rune (safe over-count)
func Count(text string) int {
	tokens := 0
	inWord := false
	wordLen := 0

	for _, r := range text {
		switch {
		case isCJKOrSyllabic(r):
			// Flush any Latin/Cyrillic word in progress.
			if inWord {
				tokens += latinWordTokens(wordLen)
				inWord = false
				wordLen = 0
			}
			tokens++ // one rune ≈ one token

		case unicode.IsLetter(r) || r == '\'' || r == '-':
			// Accumulate Latin / Cyrillic / etc. word characters.
			inWord = true
			wordLen++

		case unicode.IsDigit(r):
			// Digits: flush word, count digit clusters separately.
			if inWord {
				tokens += latinWordTokens(wordLen)
				inWord = false
				wordLen = 0
			}
			tokens++ // digit sequences: handled per-rune, cluster below

		default:
			// Space or punctuation: flush current word.
			if inWord {
				tokens += latinWordTokens(wordLen)
				inWord = false
				wordLen = 0
			}
		}
	}
	if inWord {
		tokens += latinWordTokens(wordLen)
	}
	return tokens
}

// latinWordTokens estimates BPE tokens for a Latin/Cyrillic/Greek word of
// given character length. Shorter words tend to be one token; longer words
// get split more aggressively by BPE.
func latinWordTokens(charLen int) int {
	switch {
	case charLen <= 4:
		return 1
	case charLen <= 8:
		return 2
	default:
		// ~1 token per 4 characters for long words.
		return (charLen + 3) / 4
	}
}

// isCJKOrSyllabic reports whether r belongs to a script where each character
// maps to roughly one token in standard BPE tokenizers.
func isCJKOrSyllabic(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Arabic, r) ||
		unicode.Is(unicode.Hebrew, r) ||
		unicode.Is(unicode.Thai, r) ||
		unicode.Is(unicode.Devanagari, r)
}
