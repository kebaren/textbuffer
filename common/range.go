package common

import "fmt"

// Range 表示编辑器中的范围
// (startLineNumber,startColumn) <= (endLineNumber,endColumn)
type Range struct {
	// StartLineNumber 范围开始的行号（从 1 开始）
	StartLineNumber int
	// StartColumn 范围开始的列号（从 1 开始）
	StartColumn int
	// EndLineNumber 范围结束的行号
	EndLineNumber int
	// EndColumn 范围结束的列号
	EndColumn int
}

// NewRange 创建一个新的范围
func NewRange(startLineNumber, startColumn, endLineNumber, endColumn int) *Range {
	if (startLineNumber > endLineNumber) || (startLineNumber == endLineNumber && startColumn > endColumn) {
		return &Range{
			StartLineNumber: endLineNumber,
			StartColumn:     endColumn,
			EndLineNumber:   startLineNumber,
			EndColumn:       startColumn,
		}
	}
	return &Range{
		StartLineNumber: startLineNumber,
		StartColumn:     startColumn,
		EndLineNumber:   endLineNumber,
		EndColumn:       endColumn,
	}
}

// IsEmpty 测试此范围是否为空
func (r *Range) IsEmpty() bool {
	return RangeIsEmpty(r)
}

// RangeIsEmpty 测试范围是否为空
func RangeIsEmpty(r *Range) bool {
	return r.StartLineNumber == r.EndLineNumber && r.StartColumn == r.EndColumn
}

// ContainsPosition 测试位置是否在此范围内
// 如果位置在边缘，将返回 true
func (r *Range) ContainsPosition(position *Position) bool {
	return RangeContainsPosition(r, position)
}

// RangeContainsPosition 测试位置是否在范围内
// 如果位置在边缘，将返回 true
func RangeContainsPosition(r *Range, position *Position) bool {
	if position.LineNumber < r.StartLineNumber || position.LineNumber > r.EndLineNumber {
		return false
	}
	if position.LineNumber == r.StartLineNumber && position.Column < r.StartColumn {
		return false
	}
	if position.LineNumber == r.EndLineNumber && position.Column > r.EndColumn {
		return false
	}
	return true
}

// ContainsRange 测试范围是否在此范围内
// 如果范围等于此范围，将返回 true
func (r *Range) ContainsRange(other *Range) bool {
	return RangeContainsRange(r, other)
}

// RangeContainsRange 测试 otherRange 是否在 range 内
// 如果范围相等，将返回 true
func RangeContainsRange(r *Range, other *Range) bool {
	if other.StartLineNumber < r.StartLineNumber || other.EndLineNumber < r.StartLineNumber {
		return false
	}
	if other.StartLineNumber > r.EndLineNumber || other.EndLineNumber > r.EndLineNumber {
		return false
	}
	if other.StartLineNumber == r.StartLineNumber && other.StartColumn < r.StartColumn {
		return false
	}
	if other.EndLineNumber == r.EndLineNumber && other.EndColumn > r.EndColumn {
		return false
	}
	return true
}

// PlusRange 两个范围的合并
// 最小位置将用作起点，最大位置将用作终点
func (r *Range) PlusRange(other *Range) *Range {
	return RangePlusRange(r, other)
}

// RangePlusRange 两个范围的合并
// 最小位置将用作起点，最大位置将用作终点
func RangePlusRange(a, b *Range) *Range {
	var startLineNumber, startColumn, endLineNumber, endColumn int

	if b.StartLineNumber < a.StartLineNumber {
		startLineNumber = b.StartLineNumber
		startColumn = b.StartColumn
	} else if b.StartLineNumber == a.StartLineNumber {
		startLineNumber = b.StartLineNumber
		startColumn = min(b.StartColumn, a.StartColumn)
	} else {
		startLineNumber = a.StartLineNumber
		startColumn = a.StartColumn
	}

	if b.EndLineNumber > a.EndLineNumber {
		endLineNumber = b.EndLineNumber
		endColumn = b.EndColumn
	} else if b.EndLineNumber == a.EndLineNumber {
		endLineNumber = b.EndLineNumber
		endColumn = max(b.EndColumn, a.EndColumn)
	} else {
		endLineNumber = a.EndLineNumber
		endColumn = a.EndColumn
	}

	return NewRange(startLineNumber, startColumn, endLineNumber, endColumn)
}

// IntersectRanges 两个范围的交集
func (r *Range) IntersectRanges(other *Range) *Range {
	return RangeIntersectRanges(r, other)
}

// RangeIntersectRanges 两个范围的交集
func RangeIntersectRanges(a, b *Range) *Range {
	resultStartLineNumber := a.StartLineNumber
	resultStartColumn := a.StartColumn
	resultEndLineNumber := a.EndLineNumber
	resultEndColumn := a.EndColumn
	otherStartLineNumber := b.StartLineNumber
	otherStartColumn := b.StartColumn
	otherEndLineNumber := b.EndLineNumber
	otherEndColumn := b.EndColumn

	if resultStartLineNumber < otherStartLineNumber {
		resultStartLineNumber = otherStartLineNumber
		resultStartColumn = otherStartColumn
	} else if resultStartLineNumber == otherStartLineNumber {
		resultStartColumn = max(resultStartColumn, otherStartColumn)
	}

	if resultEndLineNumber > otherEndLineNumber {
		resultEndLineNumber = otherEndLineNumber
		resultEndColumn = otherEndColumn
	} else if resultEndLineNumber == otherEndLineNumber {
		resultEndColumn = min(resultEndColumn, otherEndColumn)
	}

	// 检查选择是否现在为空
	if resultStartLineNumber > resultEndLineNumber {
		return nil
	}
	if resultStartLineNumber == resultEndLineNumber && resultStartColumn > resultEndColumn {
		return nil
	}
	return NewRange(resultStartLineNumber, resultStartColumn, resultEndLineNumber, resultEndColumn)
}

// EqualsRange 测试此范围是否等于其他范围
func (r *Range) EqualsRange(other *Range) bool {
	return RangeEqualsRange(r, other)
}

// RangeEqualsRange 测试范围 a 是否等于 b
func RangeEqualsRange(a, b *Range) bool {
	return a != nil && b != nil &&
		a.StartLineNumber == b.StartLineNumber &&
		a.StartColumn == b.StartColumn &&
		a.EndLineNumber == b.EndLineNumber &&
		a.EndColumn == b.EndColumn
}

// GetEndPosition 返回结束位置（将在开始位置之后或等于开始位置）
func (r *Range) GetEndPosition() *Position {
	return NewPosition(r.EndLineNumber, r.EndColumn)
}

// GetStartPosition 返回开始位置（将在结束位置之前或等于结束位置）
func (r *Range) GetStartPosition() *Position {
	return NewPosition(r.StartLineNumber, r.StartColumn)
}

// String 转换为人类可读的表示形式
func (r *Range) String() string {
	return fmt.Sprintf("[%d,%d -> %d,%d]", r.StartLineNumber, r.StartColumn, r.EndLineNumber, r.EndColumn)
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
