package buffer

// 常用值
const (
	// AverageBufferSize 平均缓冲区大小
	AverageBufferSize = 65536
)

// StringBuffer 字符串缓冲区
type StringBuffer struct {
	// Buffer 缓冲区
	Buffer string
	// LineStarts 行起始位置
	LineStarts []int
}

// NewStringBuffer 创建一个新的字符串缓冲区
func NewStringBuffer(buffer string, lineStarts []int) *StringBuffer {
	return &StringBuffer{
		Buffer:     buffer,
		LineStarts: lineStarts,
	}
}
