package buffer

// LineStarts 行起始位置
type LineStarts struct {
	// Offsets 偏移量
	Offsets []int
	// CRCount 回车符计数
	CRCount int
	// LFCount 换行符计数
	LFCount int
	// CRLFCount 回车换行符计数
	CRLFCount int
	// IsBasicASCII 是否是基本 ASCII
	IsBasicASCII bool
}

// NewLineStarts 创建一个新的行起始位置
func NewLineStarts() *LineStarts {
	return &LineStarts{
		Offsets:      []int{0},
		CRCount:      0,
		LFCount:      0,
		CRLFCount:    0,
		IsBasicASCII: true,
	}
}

// CreateLineStartsFast 快速创建行起始位置
func CreateLineStartsFast(text string, isBasicASCII bool) []int {
	result := []int{0}
	for i, length := 0, len(text); i < length; i++ {
		ch := text[i]
		if ch == '\r' {
			if i+1 < length && text[i+1] == '\n' {
				i++
			}
			result = append(result, i+1)
		} else if ch == '\n' {
			result = append(result, i+1)
		}
	}
	return result
}

// CreateLineStarts 创建行起始位置
func CreateLineStarts(text string) *LineStarts {
	result := NewLineStarts()
	result.IsBasicASCII = true

	for i, length := 0, len(text); i < length; i++ {
		ch := rune(text[i])

		if !IsBasicASCII(ch) {
			result.IsBasicASCII = false
		}

		if ch == '\r' {
			result.CRCount++
			if i+1 < length && text[i+1] == '\n' {
				result.CRLFCount++
				i++
			}
			result.Offsets = append(result.Offsets, i+1)
		} else if ch == '\n' {
			result.LFCount++
			result.Offsets = append(result.Offsets, i+1)
		}
	}

	return result
}

// IsBasicASCII 判断是否是基本 ASCII 字符
func IsBasicASCII(ch rune) bool {
	return ch >= 0 && ch <= 127
}
