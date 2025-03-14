package common

import "fmt"

// Position 表示编辑器中的位置
// lineNumber 从 1 开始，column 从 1 开始
type Position struct {
	// LineNumber 行号（从 1 开始）
	LineNumber int
	// Column 列号（从 1 开始）
	Column int
}

// NewPosition 创建一个新的位置
func NewPosition(lineNumber, column int) *Position {
	return &Position{
		LineNumber: lineNumber,
		Column:     column,
	}
}

// With 从当前位置创建一个新位置
func (p *Position) With(newLineNumber, newColumn int) *Position {
	if newLineNumber == p.LineNumber && newColumn == p.Column {
		return p
	}
	return NewPosition(newLineNumber, newColumn)
}

// Delta 从当前位置派生一个新位置
func (p *Position) Delta(deltaLineNumber, deltaColumn int) *Position {
	return p.With(p.LineNumber+deltaLineNumber, p.Column+deltaColumn)
}

// Equals 测试此位置是否等于其他位置
func (p *Position) Equals(other *Position) bool {
	return PositionEquals(p, other)
}

// PositionEquals 测试位置 a 是否等于位置 b
func PositionEquals(a, b *Position) bool {
	if a == nil && b == nil {
		return true
	}
	return a != nil && b != nil && a.LineNumber == b.LineNumber && a.Column == b.Column
}

// IsBefore 测试此位置是否在其他位置之前
// 如果两个位置相等，结果将为 false
func (p *Position) IsBefore(other *Position) bool {
	return PositionIsBefore(p, other)
}

// PositionIsBefore 测试位置 a 是否在位置 b 之前
// 如果两个位置相等，结果将为 false
func PositionIsBefore(a, b *Position) bool {
	if a.LineNumber < b.LineNumber {
		return true
	}
	if b.LineNumber < a.LineNumber {
		return false
	}
	return a.Column < b.Column
}

// IsBeforeOrEqual 测试此位置是否在其他位置之前或等于其他位置
func (p *Position) IsBeforeOrEqual(other *Position) bool {
	return PositionIsBeforeOrEqual(p, other)
}

// PositionIsBeforeOrEqual 测试位置 a 是否在位置 b 之前或等于位置 b
func PositionIsBeforeOrEqual(a, b *Position) bool {
	if a.LineNumber < b.LineNumber {
		return true
	}
	if b.LineNumber < a.LineNumber {
		return false
	}
	return a.Column <= b.Column
}

// Compare 比较两个位置，用于排序
func PositionCompare(a, b *Position) int {
	if a.LineNumber == b.LineNumber {
		return a.Column - b.Column
	}
	return a.LineNumber - b.LineNumber
}

// Clone 克隆此位置
func (p *Position) Clone() *Position {
	return NewPosition(p.LineNumber, p.Column)
}

// String 转换为人类可读的表示形式
func (p *Position) String() string {
	return fmt.Sprintf("(%d,%d)", p.LineNumber, p.Column)
}
