package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCountEmptyString returns 0 for empty input.
func TestCountEmptyString(t *testing.T) {
	result := Count("")
	require.Equal(t, 0, result)
}

// TestCountSingleWords counts Latin words correctly.
func TestCountSingleWords(t *testing.T) {
	tests := []struct {
		word     string
		expected int
	}{
		{"hello", 2},        // 5 chars → 2 tokens
		{"world", 2},        // 5 chars → 2 tokens
		{"configure", 3},    // 9 chars → 3 tokens
		{"optimization", 3}, // 12 chars → 3 tokens
		{"a", 1},            // 1 char → 1 token
		{"at", 1},           // 2 chars → 1 token
		{"test", 1},         // 4 chars → 1 token
		{"testing", 2},      // 7 chars → 2 tokens
	}

	for _, tt := range tests {
		result := Count(tt.word)
		require.Equal(t, tt.expected, result, "Count(%q)", tt.word)
	}
}

// TestCountMultipleWords counts word count × 1.3 heuristic.
func TestCountMultipleWords(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int // conservative minimum
		maxTokens int // generous maximum
	}{
		{"hello world", 4, 5},
		{"the quick brown fox", 5, 7},
		{"what is your name", 4, 8},
		{"python programming language", 4, 7},
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.GreaterOrEqual(t, result, tt.minTokens, "Count(%q) too low", tt.text)
		require.LessOrEqual(t, result, tt.maxTokens, "Count(%q) too high", tt.text)
	}
}

// TestCountCJK handles Chinese/Japanese/Korean characters.
func TestCountCJK(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"你好", 2},    // Chinese: 2 chars → 2 tokens
		{"こんにちは", 5}, // Japanese: 5 chars → 5 tokens
		{"안녕하세요", 5}, // Korean: 5 chars → 5 tokens
		{"中文测试", 4},  // Mixed Chinese → 4 tokens
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.Equal(t, tt.expected, result, "Count(%q)", tt.text)
	}
}

// TestCountCyrillic handles Russian/Cyrillic text.
func TestCountCyrillic(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
		maxTokens int
	}{
		{"привет", 1, 2},       // Russian word → 1-2 tokens
		{"hello привет", 2, 4}, // Mixed Latin+Cyrillic
		{"Москва", 1, 2},       // City name
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.GreaterOrEqual(t, result, tt.minTokens, "Count(%q) too low", tt.text)
		require.LessOrEqual(t, result, tt.maxTokens, "Count(%q) too high", tt.text)
	}
}

// TestCountArabic handles Arabic script.
func TestCountArabic(t *testing.T) {
	// Arabic text: each character ≈ 1 token
	text := "السلام عليكم"
	result := Count(text)
	require.Greater(t, result, 0)
}

// TestCountNumbers handles numeric sequences.
func TestCountNumbers(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
	}{
		{"123", 1},
		{"1234567890", 1},
		{"the year 2024", 3},
		{"version 3.14.159", 4},
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.GreaterOrEqual(t, result, tt.minTokens, "Count(%q)", tt.text)
	}
}

// TestCountPunctuation handles punctuation and special characters.
func TestCountPunctuation(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
	}{
		{"Hello, world!", 2},
		{"What's up?", 2},
		{"self-driving cars", 2},
		{"email@example.com", 1},
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.GreaterOrEqual(t, result, tt.minTokens, "Count(%q)", tt.text)
	}
}

// TestCountWhitespace handles various whitespace.
func TestCountWhitespace(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"   ", 0},            // Only spaces
		{"hello  world", 4},   // Double spaces
		{"hello\nworld", 4},   // Newline
		{"hello\tworld", 4},   // Tab
		{"hello \n world", 4}, // Mixed whitespace
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.Equal(t, tt.expected, result, "Count(%q)", tt.text)
	}
}

// TestCountLongDocument counts a realistic document.
func TestCountLongDocument(t *testing.T) {
	doc := `This is a comprehensive document about tokenization.
Tokens are the basic units that language models understand.
Different scripts require different counting strategies.
Latin text uses a 1.3x multiplier for BPE overhead.
CJK text counts 1 token per character.
This approach handles all Unicode scripts correctly.`

	result := Count(doc)
	require.Greater(t, result, 50, "document should have many tokens")
	require.Less(t, result, 200, "document should not have too many tokens")
}

// TestCountHyphenatedWords handles hyphenated words correctly.
func TestCountHyphenatedWords(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
	}{
		{"self-aware", 1},
		{"state-of-the-art", 2},
		{"end-to-end", 2},
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.GreaterOrEqual(t, result, tt.minTokens, "Count(%q)", tt.text)
	}
}

// TestCountApostrophes handles apostrophes in contractions.
func TestCountApostrophes(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
	}{
		{"don't", 1},
		{"it's easy", 2},
		{"I'll help", 2},
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.GreaterOrEqual(t, result, tt.minTokens, "Count(%q)", tt.text)
	}
}

// TestCountConservativeOverCount verifies conservative over-counting.
func TestCountConservativeOverCount(t *testing.T) {
	// These tests ensure we don't under-count, which would break chunk splitting
	testCases := []struct {
		text      string
		maxChars  int
		maxTokens int
	}{
		// A 500-char chunk should not exceed 200 tokens
		{genString(500), 500, 200},
		// A 1000-char chunk should not exceed 400 tokens
		{genString(1000), 1000, 400},
	}

	for _, tc := range testCases {
		result := Count(tc.text)
		require.LessOrEqual(t, result, tc.maxTokens, "over-counting limit for %d chars", len(tc.text))
	}
}

// Helper function to generate a test string
func genString(length int) string {
	word := "hello world "
	result := ""
	for len(result) < length {
		result += word
	}
	return result[:length]
}

// TestLatinWordTokensEdgeCases tests the word token calculation edge cases.
func TestLatinWordTokensEdgeCases(t *testing.T) {
	tests := []struct {
		charLen  int
		expected int
	}{
		{1, 1},
		{4, 1},
		{5, 2},
		{8, 2},
		{9, 3},
		{12, 3},
		{16, 4},
	}

	for _, tt := range tests {
		result := latinWordTokens(tt.charLen)
		require.Equal(t, tt.expected, result, "latinWordTokens(%d)", tt.charLen)
	}
}

// TestIsCJKOrSyllabic verifies script detection.
func TestIsCJKOrSyllabic(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'你', true},  // Han
		{'あ', true},  // Hiragana
		{'カ', true},  // Katakana
		{'한', true},  // Hangul
		{'ا', true},  // Arabic
		{'א', true},  // Hebrew
		{'ด', true},  // Thai
		{'अ', true},  // Devanagari
		{'a', false}, // Latin
		{'б', false}, // Cyrillic
		{'1', false}, // Digit
		{' ', false}, // Space
	}

	for _, tt := range tests {
		result := isCJKOrSyllabic(tt.r)
		require.Equal(t, tt.expected, result, "isCJKOrSyllabic(%U)", tt.r)
	}
}

// TestCountMixedScripts combines multiple scripts.
func TestCountMixedScripts(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
	}{
		{"hello 你好", 2},
		{"test اختبار", 2},
		{"english русский", 2},
		{"mixed 混合 text", 2},
	}

	for _, tt := range tests {
		result := Count(tt.text)
		require.GreaterOrEqual(t, result, tt.minTokens, "Count(%q)", tt.text)
	}
}
