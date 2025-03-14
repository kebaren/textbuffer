package buffer

import "strings"

// DefaultEndOfLine 默认换行符类型
type DefaultEndOfLine int

const (
	// LF 使用换行符 (\n) 作为行尾字符
	LF DefaultEndOfLine = 1
	// CRLF 使用回车换行符 (\r\n) 作为行尾字符
	CRLF DefaultEndOfLine = 2
)

// UTF8BOMCharacter UTF-8 BOM 字符
const UTF8BOMCharacter = "\uFEFF" // UTF8_BOM

// StartsWithUTF8BOM 检查字符串是否以 UTF-8 BOM 开头
func StartsWithUTF8BOM(str string) bool {
	return len(str) > 0 && str[0] == byte(0xEF) && str[1] == byte(0xBB) && str[2] == byte(0xBF)
}

// PieceTreeTextBufferFactory 片段树文本缓冲区工厂
type PieceTreeTextBufferFactory struct {
	chunks       []*StringBuffer
	bom          string
	cr           int
	lf           int
	crlf         int
	normalizeEOL bool
}

// NewPieceTreeTextBufferFactory 创建一个新的片段树文本缓冲区工厂
func NewPieceTreeTextBufferFactory(chunks []*StringBuffer, bom string, cr, lf, crlf int, normalizeEOL bool) *PieceTreeTextBufferFactory {
	return &PieceTreeTextBufferFactory{
		chunks:       chunks,
		bom:          bom,
		cr:           cr,
		lf:           lf,
		crlf:         crlf,
		normalizeEOL: normalizeEOL,
	}
}

// GetEOL 获取换行符
func (f *PieceTreeTextBufferFactory) GetEOL(defaultEOL DefaultEndOfLine) string {
	totalEOLCount := f.cr + f.lf + f.crlf
	totalCRCount := f.cr + f.crlf
	if totalEOLCount == 0 {
		// 这是一个空文件或只有一行的文件
		if defaultEOL == LF {
			return "\n"
		}
		return "\r\n"
	}
	if totalCRCount > totalEOLCount/2 {
		// 超过一半的文件包含 \r\n 结尾的行
		return "\r\n"
	}
	// 至少有一行以 \n 结尾
	return "\n"
}

// Create 创建片段树
func (f *PieceTreeTextBufferFactory) Create(defaultEOL DefaultEndOfLine) *PieceTreeBase {
	eol := f.GetEOL(defaultEOL)
	chunks := f.chunks

	if f.normalizeEOL &&
		((eol == "\r\n" && (f.cr > 0 || f.lf > 0)) ||
			(eol == "\n" && (f.cr > 0 || f.crlf > 0))) {
		// 规范化片段
		for i, length := 0, len(chunks); i < length; i++ {
			re := strings.NewReplacer("\r\n", eol, "\r", eol, "\n", eol)
			str := re.Replace(chunks[i].Buffer)
			newLineStarts := CreateLineStartsFast(str, true)
			chunks[i] = NewStringBuffer(str, newLineStarts)
		}
	}

	return NewPieceTreeBase(chunks, eol, f.normalizeEOL)
}

// GetFirstLineText 获取第一行文本
func (f *PieceTreeTextBufferFactory) GetFirstLineText(lengthLimit int) string {
	if len(f.chunks) == 0 || len(f.chunks[0].Buffer) == 0 {
		return ""
	}

	// 获取第一个缓冲区的前 lengthLimit 个字符
	str := f.chunks[0].Buffer
	if len(str) > lengthLimit {
		str = str[:lengthLimit]
	}

	// 分割成行并返回第一行
	lines := strings.Split(str, "\n")
	firstLine := lines[0]

	// 如果第一行以 \r 结尾，去掉它
	if len(firstLine) > 0 && firstLine[len(firstLine)-1] == '\r' {
		firstLine = firstLine[:len(firstLine)-1]
	}

	return firstLine
}

// PieceTreeTextBufferBuilder 片段树文本缓冲区构建器
type PieceTreeTextBufferBuilder struct {
	chunks          []*StringBuffer
	BOM             string
	hasPreviousChar bool
	previousChar    rune
	tmpLineStarts   []int
	cr              int
	lf              int
	crlf            int
}

// NewPieceTreeTextBufferBuilder 创建一个新的片段树文本缓冲区构建器
func NewPieceTreeTextBufferBuilder() *PieceTreeTextBufferBuilder {
	return &PieceTreeTextBufferBuilder{
		chunks:          make([]*StringBuffer, 0),
		BOM:             "",
		hasPreviousChar: false,
		previousChar:    0,
		tmpLineStarts:   make([]int, 0),
		cr:              0,
		lf:              0,
		crlf:            0,
	}
}

// AcceptChunk 接受一个文本块
func (b *PieceTreeTextBufferBuilder) AcceptChunk(chunk string) {
	if len(chunk) == 0 {
		return
	}

	if len(b.chunks) == 0 {
		if StartsWithUTF8BOM(chunk) {
			b.BOM = UTF8BOMCharacter
			chunk = chunk[1:]
		}
	}

	lastChar := rune(chunk[len(chunk)-1])
	if lastChar == '\r' || (lastChar >= 0xD800 && lastChar <= 0xDBFF) {
		// 最后一个字符是 \r 或高代理项 => 保留它
		b.acceptChunk1(chunk[:len(chunk)-1], false)
		b.hasPreviousChar = true
		b.previousChar = lastChar
	} else {
		b.acceptChunk1(chunk, false)
		b.hasPreviousChar = false
		b.previousChar = lastChar
	}
}

// acceptChunk1 接受一个文本块（内部方法）
func (b *PieceTreeTextBufferBuilder) acceptChunk1(chunk string, allowEmptyStrings bool) {
	if !allowEmptyStrings && len(chunk) == 0 {
		// 没有要做的事情
		return
	}

	if b.hasPreviousChar {
		b.acceptChunk2(string(b.previousChar) + chunk)
	} else {
		b.acceptChunk2(chunk)
	}
}

// acceptChunk2 接受一个文本块（内部方法）
func (b *PieceTreeTextBufferBuilder) acceptChunk2(chunk string) {
	lineStarts := CreateLineStarts(chunk)

	b.chunks = append(b.chunks, NewStringBuffer(chunk, lineStarts.Offsets))
	b.cr += lineStarts.CRCount
	b.lf += lineStarts.LFCount
	b.crlf += lineStarts.CRLFCount
}

// Finish 完成构建
func (b *PieceTreeTextBufferBuilder) Finish(normalizeEOL bool) *PieceTreeTextBufferFactory {
	b.finish()
	return NewPieceTreeTextBufferFactory(
		b.chunks,
		b.BOM,
		b.cr,
		b.lf,
		b.crlf,
		normalizeEOL,
	)
}

// finish 完成构建（内部方法）
func (b *PieceTreeTextBufferBuilder) finish() {
	if len(b.chunks) == 0 {
		b.acceptChunk1("", true)
	}

	if b.hasPreviousChar {
		b.hasPreviousChar = false
		// 重新创建最后一个块
		lastChunk := b.chunks[len(b.chunks)-1]
		lastChunk.Buffer += string(b.previousChar)
		newLineStarts := CreateLineStartsFast(lastChunk.Buffer, true)
		lastChunk.LineStarts = newLineStarts
		if b.previousChar == '\r' {
			b.cr++
		}
	}
}
