package buffer

import (
	"testing"
)

// TestNewPieceTree 测试创建新的PieceTree
func TestNewPieceTree(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{"空字符串", "", ""},
		{"简单文本", "Hello, World!", "Hello, World!"},
		{"多行文本", "Line 1\nLine 2\nLine 3", "Line 1\nLine 2\nLine 3"},
		{"CRLF文本", "Line 1\r\nLine 2\r\nLine 3", "Line 1\r\nLine 2\r\nLine 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPieceTree(tt.text)
			if tree == nil {
				t.Error("NewPieceTree returned nil")
				return
			}
			content := tree.GetLinesContent()
			if len(content) == 0 && tt.expected != "" {
				t.Error("GetLinesContent returned empty slice for non-empty text")
				return
			}
			actual := ""
			for i, line := range content {
				if i > 0 {
					actual += "\n"
				}
				actual += line
			}
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestInsert 测试插入操作
func TestInsert(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		offset   int
		text     string
		expected string
	}{
		{"在开头插入", "World", 0, "Hello, ", "Hello, World"},
		{"在中间插入", "Hello World", 6, ", ", "Hello, World"},
		{"在结尾插入", "Hello", 5, ", World", "Hello, World"},
		{"插入空字符串", "Hello", 5, "", "Hello"},
		{"插入换行符", "Hello", 5, "\nWorld", "Hello\nWorld"},
		{"插入CRLF", "Hello", 5, "\r\nWorld", "Hello\r\nWorld"},
		{"越界插入", "Hello", 10, ", World", "Hello, World"},
		{"负偏移插入", "Hello", -1, ", World", ", WorldHello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPieceTree(tt.initial)
			tree.Insert(tt.offset, tt.text, true)
			content := tree.GetLinesContent()
			actual := ""
			for i, line := range content {
				if i > 0 {
					actual += "\n"
				}
				actual += line
			}
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestDelete 测试删除操作
func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		offset   int
		length   int
		expected string
	}{
		{"删除开头", "Hello, World", 0, 7, "World"},
		{"删除中间", "Hello, World", 6, 2, "Hello World"},
		{"删除结尾", "Hello, World", 7, 5, "Hello, "},
		{"删除全部", "Hello, World", 0, 12, ""},
		{"删除空", "Hello", 2, 0, "Hello"},
		{"越界删除", "Hello", 2, 10, "He"},
		{"负长度删除", "Hello", 2, -1, "Hello"},
		{"负偏移删除", "Hello", -1, 2, "llo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPieceTree(tt.initial)
			tree.Delete(tt.offset, tt.length)
			content := tree.GetLinesContent()
			actual := ""
			for i, line := range content {
				if i > 0 {
					actual += "\n"
				}
				actual += line
			}
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestGetLineContent 测试获取行内容
func TestGetLineContent(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		lineNum  int
		expected string
	}{
		{"第一行", "Line 1\nLine 2\nLine 3", 1, "Line 1"},
		{"中间行", "Line 1\nLine 2\nLine 3", 2, "Line 2"},
		{"最后一行", "Line 1\nLine 2\nLine 3", 3, "Line 3"},
		{"空行", "\n\n", 2, ""},
		{"越界行号", "Line 1\nLine 2", 3, ""},
		{"负行号", "Line 1\nLine 2", 0, ""},
		{"CRLF行", "Line 1\r\nLine 2\r\nLine 3", 2, "Line 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPieceTree(tt.text)
			actual := tree.GetLineContent(tt.lineNum)
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestGetLineCount 测试获取行数
func TestGetLineCount(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"空文本", "", 1},
		{"单行文本", "Hello, World", 1},
		{"多行文本", "Line 1\nLine 2\nLine 3", 3},
		{"CRLF文本", "Line 1\r\nLine 2\r\nLine 3", 3},
		{"混合换行符", "Line 1\nLine 2\r\nLine 3", 3},
		{"空行", "\n\n", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPieceTree(tt.text)
			actual := tree.GetLineCount()
			if actual != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}

// TestGetLength 测试获取文本长度
func TestGetLength(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"空文本", "", 0},
		{"简单文本", "Hello", 5},
		{"多行文本", "Line 1\nLine 2", 11},
		{"CRLF文本", "Line 1\r\nLine 2", 12},
		{"特殊字符", "Hello\n\tWorld", 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPieceTree(tt.text)
			actual := tree.GetLength()
			if actual != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}

// TestGetPositionAt 测试获取位置
func TestGetPositionAt(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		offset   int
		expected struct {
			line   int
			column int
		}
	}{
		{"开始位置", "Hello", 0, struct {
			line   int
			column int
		}{1, 1}},
		{"中间位置", "Hello", 2, struct {
			line   int
			column int
		}{1, 3}},
		{"结束位置", "Hello", 5, struct {
			line   int
			column int
		}{1, 6}},
		{"换行位置", "Hello\nWorld", 5, struct {
			line   int
			column int
		}{2, 1}},
		{"越界位置", "Hello", 10, struct {
			line   int
			column int
		}{1, 6}},
		{"负偏移", "Hello", -1, struct {
			line   int
			column int
		}{1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPieceTree(tt.text)
			pos := tree.GetPositionAt(tt.offset)
			if pos.LineNumber != tt.expected.line || pos.Column != tt.expected.column {
				t.Errorf("expected line %d, column %d; got line %d, column %d",
					tt.expected.line, tt.expected.column, pos.LineNumber, pos.Column)
			}
		})
	}
}

// TestGetOffsetAt 测试获取偏移量
func TestGetOffsetAt(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		line     int
		column   int
		expected int
	}{
		{"开始位置", "Hello", 1, 1, 0},
		{"中间位置", "Hello", 1, 3, 2},
		{"结束位置", "Hello", 1, 6, 5},
		{"换行位置", "Hello\nWorld", 2, 1, 6},
		{"越界行", "Hello", 2, 1, 5},
		{"越界列", "Hello", 1, 10, 5},
		{"负行号", "Hello", 0, 1, 0},
		{"负列号", "Hello", 1, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewPieceTree(tt.text)
			actual := tree.GetOffsetAt(tt.line, tt.column)
			if actual != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}

// TestComplexOperations 测试复杂操作组合
func TestComplexOperations(t *testing.T) {
	tree := NewPieceTree("Hello")

	// 测试插入和删除的组合
	tree.Insert(5, ", World", true)
	if tree.GetLength() != 12 {
		t.Errorf("expected length 12, got %d", tree.GetLength())
	}

	tree.Delete(5, 2)
	if tree.GetLength() != 10 {
		t.Errorf("expected length 10, got %d", tree.GetLength())
	}

	// 测试多行操作
	tree.Insert(10, "\nNew Line", true)
	if tree.GetLineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", tree.GetLineCount())
	}

	// 测试位置转换
	pos := tree.GetPositionAt(10)
	if pos.LineNumber != 2 || pos.Column != 1 {
		t.Errorf("expected line 2, column 1; got line %d, column %d", pos.LineNumber, pos.Column)
	}

	offset := tree.GetOffsetAt(2, 1)
	if offset != 10 {
		t.Errorf("expected offset 10, got %d", offset)
	}

	// 测试行内容
	line1 := tree.GetLineContent(1)
	if line1 != "Hello, Worl" {
		t.Errorf("expected 'Hello, Worl', got %q", line1)
	}

	line2 := tree.GetLineContent(2)
	if line2 != "New Line" {
		t.Errorf("expected 'New Line', got %q", line2)
	}
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	// 测试空树
	tree := NewPieceTree("")
	if tree.GetLength() != 0 {
		t.Error("empty tree should have length 0")
	}
	if tree.GetLineCount() != 1 {
		t.Error("empty tree should have 1 line")
	}

	// 测试大量文本
	longText := ""
	for i := 0; i < 1000; i++ {
		longText += "Line " + string(rune(i%10+'0')) + "\n"
	}
	tree = NewPieceTree(longText)
	if tree.GetLineCount() != 1001 {
		t.Errorf("expected 1001 lines, got %d", tree.GetLineCount())
	}

	// 测试特殊字符
	specialChars := "Hello\x00World\n\t\r\n"
	tree = NewPieceTree(specialChars)
	if tree.GetLength() != len(specialChars) {
		t.Errorf("expected length %d, got %d", len(specialChars), tree.GetLength())
	}

	// 测试Unicode字符
	unicodeText := "Hello 世界\n你好 World"
	tree = NewPieceTree(unicodeText)
	if tree.GetLength() != len(unicodeText) {
		t.Errorf("expected length %d, got %d", len(unicodeText), tree.GetLength())
	}
	if tree.GetLineCount() != 2 {
		t.Error("expected 2 lines")
	}
}
