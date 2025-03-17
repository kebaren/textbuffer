package buffer

import (
	"regexp"
	"strings"

	"github.com/kebaren/lemon/pkg/common"
)

// PieceTreeBase 片段树基础结构
type PieceTreeBase struct {
	// Root 根节点
	Root *TreeNode
	// buffers 缓冲区数组，0 是变更缓冲区，其他是只读原始缓冲区
	buffers []*StringBuffer
	// lineCnt 行数
	lineCnt int
	// length 长度
	length int
	// EOL 换行符
	EOL string
	// EOLLength 换行符长度
	EOLLength int
	// EOLNormalized 是否已规范化换行符
	EOLNormalized bool
	// lastChangeBufferPos 最后变更缓冲区位置
	lastChangeBufferPos BufferCursor
	// searchCache 搜索缓存
	searchCache *PieceTreeSearchCache
	// lastVisitedLine 最后访问的行
	lastVisitedLine struct {
		LineNumber int
		Value      string
	}
}

// NewPieceTreeBase 创建一个新的片段树基础结构
func NewPieceTreeBase(chunks []*StringBuffer, eol string, eolNormalized bool) *PieceTreeBase {
	tree := &PieceTreeBase{}
	tree.Create(chunks, eol, eolNormalized)
	return tree
}

// Create 创建片段树
func (t *PieceTreeBase) Create(chunks []*StringBuffer, eol string, eolNormalized bool) {
	t.buffers = []*StringBuffer{
		NewStringBuffer("", []int{0}),
	}
	t.lastChangeBufferPos = BufferCursor{Line: 0, Column: 0}
	t.Root = SENTINEL
	t.lineCnt = 1
	t.length = 0
	t.EOL = eol
	t.EOLLength = len(eol)
	t.EOLNormalized = eolNormalized

	var lastNode *TreeNode = nil
	for i := 0; i < len(chunks); i++ {
		if len(chunks[i].Buffer) > 0 {
			if chunks[i].LineStarts == nil {
				chunks[i].LineStarts = CreateLineStartsFast(chunks[i].Buffer, true)
			}

			piece := NewPiece(
				i+1,
				BufferCursor{Line: 0, Column: 0},
				BufferCursor{
					Line:   len(chunks[i].LineStarts) - 1,
					Column: len(chunks[i].Buffer) - chunks[i].LineStarts[len(chunks[i].LineStarts)-1],
				},
				len(chunks[i].LineStarts)-1,
				len(chunks[i].Buffer),
			)
			t.buffers = append(t.buffers, chunks[i])
			lastNode = t.RbInsertRight(lastNode, piece)
		}
	}

	t.searchCache = NewPieceTreeSearchCache(1)
	t.lastVisitedLine.LineNumber = 0
	t.lastVisitedLine.Value = ""
	t.ComputeBufferMetadata()
}

// NormalizeEOL 规范化换行符
func (t *PieceTreeBase) NormalizeEOL(eol string) {
	averageBufferSize := AverageBufferSize
	min := averageBufferSize - averageBufferSize/3
	max := min * 2

	tempChunk := ""
	tempChunkLen := 0
	chunks := make([]*StringBuffer, 0)

	t.Iterate(t.Root, func(node *TreeNode) bool {
		str := t.GetNodeContent(node)
		length := len(str)
		if tempChunkLen <= min || tempChunkLen+length < max {
			tempChunk += str
			tempChunkLen += length
			return true
		}

		// 刷新
		re := regexp.MustCompile(`\r\n|\r|\n`)
		text := re.ReplaceAllString(tempChunk, eol)
		chunks = append(chunks, NewStringBuffer(text, CreateLineStartsFast(text, true)))
		tempChunk = str
		tempChunkLen = length
		return true
	})

	if tempChunkLen > 0 {
		re := regexp.MustCompile(`\r\n|\r|\n`)
		text := re.ReplaceAllString(tempChunk, eol)
		chunks = append(chunks, NewStringBuffer(text, CreateLineStartsFast(text, true)))
	}

	t.Create(chunks, eol, true)
}

// GetEOL 获取换行符
func (t *PieceTreeBase) GetEOL() string {
	return t.EOL
}

// SetEOL 设置换行符
func (t *PieceTreeBase) SetEOL(newEOL string) {
	t.EOL = newEOL
	t.EOLLength = len(newEOL)
	t.NormalizeEOL(newEOL)
}

// CreateSnapshot 创建快照
func (t *PieceTreeBase) CreateSnapshot(BOM string) ITextSnapshot {
	return NewPieceTreeSnapshot(t, BOM)
}

// Equal 比较两个片段树是否相等
func (t *PieceTreeBase) Equal(other *PieceTreeBase) bool {
	// 比较长度
	if t.length != other.length || t.lineCnt != other.lineCnt {
		return false
	}

	// 比较内容
	snapshot1 := t.CreateSnapshot("")
	content1 := snapshot1.Read()

	snapshot2 := other.CreateSnapshot("")
	content2 := snapshot2.Read()

	return content1 == content2
}

// GetLength 获取长度
func (t *PieceTreeBase) GetLength() int {
	return t.length
}

// GetLineCount 获取行数
func (t *PieceTreeBase) GetLineCount() int {
	return t.lineCnt
}

// Iterate 遍历
func (t *PieceTreeBase) Iterate(node *TreeNode, callback func(node *TreeNode) bool) bool {
	if node == SENTINEL {
		return callback(SENTINEL)
	}

	leftRet := t.Iterate(node.Left, callback)
	if !leftRet {
		return leftRet
	}

	return callback(node) && t.Iterate(node.Right, callback)
}

// GetNodeContent 获取节点内容
func (t *PieceTreeBase) GetNodeContent(node *TreeNode) string {
	if node == nil || node == SENTINEL {
		return ""
	}

	piece := node.Piece
	buffer := t.buffers[piece.BufferIndex].Buffer
	startOffset := t.OffsetInBuffer(piece.BufferIndex, piece.Start)
	endOffset := t.OffsetInBuffer(piece.BufferIndex, piece.End)

	if startOffset >= len(buffer) || endOffset > len(buffer) || startOffset > endOffset {
		return ""
	}

	return buffer[startOffset:endOffset]
}

// GetPieceContent 获取片段内容
func (t *PieceTreeBase) GetPieceContent(piece Piece) string {
	buffer := t.buffers[piece.BufferIndex]
	startOffset := t.OffsetInBuffer(piece.BufferIndex, piece.Start)
	endOffset := t.OffsetInBuffer(piece.BufferIndex, piece.End)
	currentContent := buffer.Buffer[startOffset:endOffset]
	return currentContent
}

// OffsetInBuffer 获取缓冲区中的偏移量
func (t *PieceTreeBase) OffsetInBuffer(bufferIndex int, cursor BufferCursor) int {
	lineStarts := t.buffers[bufferIndex].LineStarts
	return lineStarts[cursor.Line] + cursor.Column
}

// NodeAt 根据偏移量获取节点位置
func (t *PieceTreeBase) NodeAt(offset int) NodePosition {
	x := t.Root
	cache := t.searchCache.Get(offset)
	if cache != nil {
		return NodePosition{
			Node:            cache.Node,
			Remainder:       offset - cache.NodeStartOffset,
			NodeStartOffset: cache.NodeStartOffset,
		}
	}

	nodeStartOffset := 0

	for x != SENTINEL {
		// 如果偏移量在左子树范围内
		if x.SizeLeft > offset {
			x = x.Left
		} else if x.SizeLeft+x.Piece.Length > offset {
			// 偏移量在当前节点范围内
			nodeStartOffset += x.SizeLeft
			ret := NodePosition{
				Node:            x,
				Remainder:       offset - x.SizeLeft,
				NodeStartOffset: nodeStartOffset,
			}
			t.searchCache.Set(CacheEntry{
				Node:                ret.Node,
				NodeStartOffset:     ret.NodeStartOffset,
				NodeStartLineNumber: ret.Remainder,
			})
			return ret
		} else {
			// 偏移量在右子树范围内
			offset -= x.SizeLeft + x.Piece.Length
			nodeStartOffset += x.SizeLeft + x.Piece.Length
			x = x.Right
		}
	}

	// 如果没有找到合适的节点，返回最后一个节点
	if t.Root != SENTINEL {
		lastNode := t.FindLastNode()
		return NodePosition{
			Node:            lastNode,
			Remainder:       lastNode.Piece.Length,
			NodeStartOffset: t.OffsetOfNode(lastNode),
		}
	}

	return NodePosition{}
}

// NodeAt2 根据行号和列号获取节点位置
func (t *PieceTreeBase) NodeAt2(lineNumber, column int) NodePosition {
	if lineNumber < 1 || lineNumber > t.GetLineCount() {
		return NodePosition{}
	}

	x := t.Root
	nodeStartOffset := 0

	// Search in tree
	for x != SENTINEL {
		if x.Left != SENTINEL && x.LFLeft >= lineNumber-1 {
			x = x.Left
		} else if x.LFLeft+x.Piece.LineFeedCnt > lineNumber-1 {
			prevAccumulatedValue := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-2)
			accumulatedValue := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-1)
			nodeStartOffset += x.SizeLeft

			return NodePosition{
				Node:            x,
				Remainder:       min(prevAccumulatedValue+column-1, accumulatedValue),
				NodeStartOffset: nodeStartOffset,
			}
		} else if x.LFLeft+x.Piece.LineFeedCnt == lineNumber-1 {
			prevAccumulatedValue := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-2)
			if prevAccumulatedValue+column-1 <= x.Piece.Length {
				return NodePosition{
					Node:            x,
					Remainder:       prevAccumulatedValue + column - 1,
					NodeStartOffset: nodeStartOffset,
				}
			} else {
				column -= x.Piece.Length - prevAccumulatedValue
				break
			}
		} else {
			lineNumber -= x.LFLeft + x.Piece.LineFeedCnt
			nodeStartOffset += x.SizeLeft + x.Piece.Length
			x = x.Right
		}
	}

	// Search in order to find the node containing position.column
	x = x.Next()
	for x != SENTINEL {
		if x.Piece.LineFeedCnt > 0 {
			accumulatedValue := t.GetAccumulatedValue(x, 0)
			nodeStartOffset := t.OffsetOfNode(x)
			return NodePosition{
				Node:            x,
				Remainder:       min(column-1, accumulatedValue),
				NodeStartOffset: nodeStartOffset,
			}
		} else {
			if x.Piece.Length >= column-1 {
				nodeStartOffset := t.OffsetOfNode(x)
				return NodePosition{
					Node:            x,
					Remainder:       column - 1,
					NodeStartOffset: nodeStartOffset,
				}
			} else {
				column -= x.Piece.Length
			}
		}

		x = x.Next()
	}

	return NodePosition{}
}

// GetValueInRange2 获取范围内的值
func (t *PieceTreeBase) GetValueInRange2(startPosition, endPosition NodePosition) string {
	// 检查节点是否为nil
	if startPosition.Node == nil || endPosition.Node == nil {
		return ""
	}

	if startPosition.Node == endPosition.Node {
		node := startPosition.Node
		buffer := t.buffers[node.Piece.BufferIndex].Buffer
		startOffset := t.OffsetInBuffer(node.Piece.BufferIndex, node.Piece.Start)
		return buffer[startOffset+startPosition.Remainder : startOffset+endPosition.Remainder]
	}

	x := startPosition.Node
	buffer := t.buffers[x.Piece.BufferIndex].Buffer
	startOffset := t.OffsetInBuffer(x.Piece.BufferIndex, x.Piece.Start)
	ret := buffer[startOffset+startPosition.Remainder : startOffset+x.Piece.Length]

	x = x.Next()
	for x != SENTINEL {
		buffer := t.buffers[x.Piece.BufferIndex].Buffer
		startOffset := t.OffsetInBuffer(x.Piece.BufferIndex, x.Piece.Start)

		if x == endPosition.Node {
			ret += buffer[startOffset : startOffset+endPosition.Remainder]
			break
		} else {
			ret += buffer[startOffset : startOffset+x.Piece.Length]
		}

		x = x.Next()
	}

	return ret
}

// RbInsertRight 在右侧插入节点
func (t *PieceTreeBase) RbInsertRight(node *TreeNode, p Piece) *TreeNode {
	z := NewTreeNode(p, Red)
	z.Left = SENTINEL
	z.Right = SENTINEL
	z.Parent = SENTINEL
	z.SizeLeft = 0
	z.LFLeft = 0

	x := t.Root

	if x == SENTINEL {
		t.Root = z
		z.Color = Black
	} else if node.Right == SENTINEL {
		node.Right = z
		z.Parent = node
	} else {
		nextNode := Leftest(node.Right)
		nextNode.Left = z
		z.Parent = nextNode
	}

	FixInsert(t, z)
	return z
}

// RbInsertLeft 在左侧插入节点
func (t *PieceTreeBase) RbInsertLeft(node *TreeNode, p Piece) *TreeNode {
	z := NewTreeNode(p, Red)
	z.Left = SENTINEL
	z.Right = SENTINEL
	z.Parent = SENTINEL
	z.SizeLeft = 0
	z.LFLeft = 0

	if t.Root == SENTINEL {
		t.Root = z
		z.Color = Black
	} else if node.Left == SENTINEL {
		node.Left = z
		z.Parent = node
	} else {
		prevNode := Righttest(node.Left)
		prevNode.Right = z
		z.Parent = prevNode
	}

	FixInsert(t, z)
	return z
}

// ComputeBufferMetadata 计算缓冲区元数据
func (t *PieceTreeBase) ComputeBufferMetadata() {
	// 直接计算所有节点的长度和换行符数量
	t.length = 0
	t.lineCnt = 1 // 初始为1，因为即使空文档也有一行

	// 收集所有节点
	nodes := make([]*TreeNode, 0)
	collectAllNodes(t.Root, &nodes)

	// 计算总长度和换行符数量
	for _, node := range nodes {
		t.length += node.Piece.Length
		t.lineCnt += node.Piece.LineFeedCnt
	}

	t.searchCache.Validate(t.length)
}

// collectAllNodes 收集所有节点
func collectAllNodes(node *TreeNode, nodes *[]*TreeNode) {
	if node == SENTINEL {
		return
	}

	collectAllNodes(node.Left, nodes)
	*nodes = append(*nodes, node)
	collectAllNodes(node.Right, nodes)
}

// DeleteNode 删除节点
func (t *PieceTreeBase) DeleteNode(node *TreeNode) {
	RbDelete(t, node)
}

// DeleteNodeTail 删除节点尾部
func (t *PieceTreeBase) DeleteNodeTail(node *TreeNode, pos BufferCursor) {
	piece := node.Piece
	originalLFCnt := piece.LineFeedCnt
	originalEndOffset := t.OffsetInBuffer(piece.BufferIndex, piece.End)

	newEnd := pos
	newEndOffset := t.OffsetInBuffer(piece.BufferIndex, newEnd)
	newLineFeedCnt := t.GetLineFeedCnt(piece.BufferIndex, piece.Start, newEnd)

	lfDelta := newLineFeedCnt - originalLFCnt
	sizeDelta := newEndOffset - originalEndOffset
	newLength := piece.Length + sizeDelta

	node.Piece = NewPiece(
		piece.BufferIndex,
		piece.Start,
		newEnd,
		newLineFeedCnt,
		newLength,
	)

	UpdateTreeMetadata(t, node, sizeDelta, lfDelta)
}

// DeleteNodeHead 删除节点头部
func (t *PieceTreeBase) DeleteNodeHead(node *TreeNode, pos BufferCursor) {
	piece := node.Piece
	originalLFCnt := piece.LineFeedCnt
	originalStartOffset := t.OffsetInBuffer(piece.BufferIndex, piece.Start)

	newStart := pos
	newLineFeedCnt := t.GetLineFeedCnt(piece.BufferIndex, newStart, piece.End)
	newStartOffset := t.OffsetInBuffer(piece.BufferIndex, newStart)
	lfDelta := newLineFeedCnt - originalLFCnt
	sizeDelta := originalStartOffset - newStartOffset
	newLength := piece.Length + sizeDelta

	node.Piece = NewPiece(
		piece.BufferIndex,
		newStart,
		piece.End,
		newLineFeedCnt,
		newLength,
	)

	UpdateTreeMetadata(t, node, sizeDelta, lfDelta)
}

// GetLineFeedCnt 获取指定范围内的换行符数量
func (t *PieceTreeBase) GetLineFeedCnt(bufferIndex int, start, end BufferCursor) int {
	// 如果start和end相同，则没有换行符
	if start.Line == end.Line && start.Column == end.Column {
		return 0
	}

	// 如果end.column为0，说明end正好在行首
	if end.Column == 0 {
		return end.Line - start.Line
	}

	lineStarts := t.buffers[bufferIndex].LineStarts
	if end.Line == len(lineStarts)-1 {
		// 说明end后面没有\n，否则会有更多的lineStart
		return end.Line - start.Line
	}

	nextLineStartOffset := lineStarts[end.Line+1]
	endOffset := lineStarts[end.Line] + end.Column
	if nextLineStartOffset > endOffset+1 {
		// end后面有超过1个字符，说明不可能是\n
		return end.Line - start.Line
	}

	// endOffset + 1 == nextLineStartOffset
	// endOffset处的字符是\n，所以我们先检查前一个字符
	// 如果endOffset处的字符是\r，end.Column应该是0，不会走到这里
	previousCharOffset := endOffset - 1 // end.Column > 0 所以这里是安全的
	buffer := t.buffers[bufferIndex].Buffer

	if buffer[previousCharOffset] == '\r' {
		return end.Line - start.Line + 1
	}
	return end.Line - start.Line
}

// PositionInBuffer 根据节点和偏移量计算缓冲区中的位置
func (t *PieceTreeBase) PositionInBuffer(node *TreeNode, remainder int) BufferCursor {
	piece := node.Piece
	bufferIndex := node.Piece.BufferIndex
	lineStarts := t.buffers[bufferIndex].LineStarts

	startOffset := lineStarts[piece.Start.Line] + piece.Start.Column
	offset := startOffset + remainder

	// 二分查找 offset 在 lineStarts 中的位置
	low := piece.Start.Line
	high := piece.End.Line

	var mid int
	var midStart int
	var midStop int

	for low <= high {
		mid = low + ((high - low) / 2)
		midStart = lineStarts[mid]

		if mid == high {
			break
		}

		midStop = lineStarts[mid+1]

		if offset < midStart {
			high = mid - 1
		} else if offset >= midStop {
			low = mid + 1
		} else {
			break
		}
	}

	return BufferCursor{
		Line:   mid,
		Column: offset - midStart,
	}
}

// GetAccumulatedValue 获取累积值
func (t *PieceTreeBase) GetAccumulatedValue(node *TreeNode, index int) int {
	if index < 0 {
		return 0
	}
	piece := node.Piece
	lineStarts := t.buffers[piece.BufferIndex].LineStarts
	expectedLineStartIndex := piece.Start.Line + index + 1
	if expectedLineStartIndex > piece.End.Line {
		return lineStarts[piece.End.Line] + piece.End.Column - lineStarts[piece.Start.Line] - piece.Start.Column
	}
	return lineStarts[expectedLineStartIndex] - lineStarts[piece.Start.Line] - piece.Start.Column
}

// OffsetOfNode 获取节点的偏移量
func (t *PieceTreeBase) OffsetOfNode(node *TreeNode) int {
	if node == nil {
		return 0
	}
	pos := node.SizeLeft
	for node != t.Root {
		if node.Parent.Right == node {
			pos += node.Parent.SizeLeft + node.Parent.Piece.Length
		}

		node = node.Parent
	}

	return pos
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetIndexOf 获取节点中指定累积值的索引和余数
func (t *PieceTreeBase) GetIndexOf(node *TreeNode, accumulatedValue int) struct {
	Index     int
	Remainder int
} {
	piece := node.Piece
	pos := t.PositionInBuffer(node, accumulatedValue)
	lineCnt := pos.Line - piece.Start.Line

	if t.OffsetInBuffer(piece.BufferIndex, piece.End)-t.OffsetInBuffer(piece.BufferIndex, piece.Start) == accumulatedValue {
		// 我们正在检查此节点的末尾，所以需要进行 CRLF 检查
		realLineCnt := t.GetLineFeedCnt(node.Piece.BufferIndex, piece.Start, pos)
		if realLineCnt != lineCnt {
			// 是的，CRLF
			return struct {
				Index     int
				Remainder int
			}{
				Index:     realLineCnt,
				Remainder: 0,
			}
		}
	}

	return struct {
		Index     int
		Remainder int
	}{
		Index:     lineCnt,
		Remainder: pos.Column,
	}
}

// GetPositionAt 获取指定偏移量的位置
func (t *PieceTreeBase) GetPositionAt(offset int) *common.Position {
	if offset > t.length {
		return common.NewPosition(t.lineCnt, 1)
	}

	pos := t.NodeAt(offset)
	if pos.Node == nil {
		return common.NewPosition(1, 1)
	}

	remainder := pos.Remainder
	node := pos.Node

	var out struct {
		Index     int
		Remainder int
	}

	if remainder == 0 {
		out = struct {
			Index     int
			Remainder int
		}{
			Index:     0,
			Remainder: 0,
		}
	} else {
		out = t.GetIndexOf(node, remainder)
	}

	lineNumber := node.LFLeft + out.Index + 1
	column := out.Remainder + 1

	return common.NewPosition(lineNumber, column)
}

// GetContentOfSubTree 获取子树的内容
func (t *PieceTreeBase) GetContentOfSubTree(node *TreeNode) string {
	str := ""

	t.Iterate(node, func(node *TreeNode) bool {
		str += t.GetNodeContent(node)
		return true
	})

	return str
}

// GetLinesRawContent 获取指定行的原始内容，包括行尾字符
func (t *PieceTreeBase) GetLineRawContent(lineNumber int, endOffset int) string {
	if lineNumber < 1 || lineNumber > t.GetLineCount() {
		return ""
	}

	x := t.Root
	var ret string

	// Try to use search cache
	if t.searchCache != nil {
		cache := t.searchCache.Get2(lineNumber)
		if cache != nil {
			x = cache.Node
			prevAccumulatedValue := t.GetAccumulatedValue(x, lineNumber-cache.NodeStartLineNumber-2)
			buffer := t.buffers[x.Piece.BufferIndex].Buffer
			startOffset := t.OffsetInBuffer(x.Piece.BufferIndex, x.Piece.Start)

			if cache.NodeStartLineNumber+x.Piece.LineFeedCnt == lineNumber {
				ret = buffer[startOffset+prevAccumulatedValue : startOffset+x.Piece.Length]
			} else {
				accumulatedValue := t.GetAccumulatedValue(x, lineNumber-cache.NodeStartLineNumber-1)
				return buffer[startOffset+prevAccumulatedValue : startOffset+accumulatedValue-endOffset]
			}
		}
	}

	nodeStartOffset := 0
	originalLineNumber := lineNumber

	// Search in tree
	for x != SENTINEL {
		if x.Left != SENTINEL && x.LFLeft >= lineNumber-1 {
			x = x.Left
		} else if x.LFLeft+x.Piece.LineFeedCnt > lineNumber-1 {
			prevAccumulatedValue := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-2)
			accumulatedValue := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-1)
			buffer := t.buffers[x.Piece.BufferIndex].Buffer
			startOffset := t.OffsetInBuffer(x.Piece.BufferIndex, x.Piece.Start)
			nodeStartOffset += x.SizeLeft

			if t.searchCache != nil {
				t.searchCache.Set(CacheEntry{
					Node:                x,
					NodeStartOffset:     nodeStartOffset,
					NodeStartLineNumber: originalLineNumber - (lineNumber - 1 - x.LFLeft),
				})
			}

			return buffer[startOffset+prevAccumulatedValue : startOffset+accumulatedValue-endOffset]
		} else if x.LFLeft+x.Piece.LineFeedCnt == lineNumber-1 {
			prevAccumulatedValue := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-2)
			buffer := t.buffers[x.Piece.BufferIndex].Buffer
			startOffset := t.OffsetInBuffer(x.Piece.BufferIndex, x.Piece.Start)

			ret = buffer[startOffset+prevAccumulatedValue : startOffset+x.Piece.Length]
			break
		} else {
			lineNumber -= x.LFLeft + x.Piece.LineFeedCnt
			nodeStartOffset += x.SizeLeft + x.Piece.Length
			x = x.Right
		}
	}

	// Search in order to find the node containing end column
	x = x.Next()
	for x != SENTINEL {
		buffer := t.buffers[x.Piece.BufferIndex].Buffer

		if x.Piece.LineFeedCnt > 0 {
			accumulatedValue := t.GetAccumulatedValue(x, 0)
			startOffset := t.OffsetInBuffer(x.Piece.BufferIndex, x.Piece.Start)

			ret += buffer[startOffset : startOffset+accumulatedValue-endOffset]
			return ret
		} else {
			startOffset := t.OffsetInBuffer(x.Piece.BufferIndex, x.Piece.Start)
			ret += buffer[startOffset : startOffset+x.Piece.Length]
		}

		x = x.Next()
	}

	return ret
}

// GetLineContent 获取指定行的内容
func (t *PieceTreeBase) GetLineContent(lineNumber int) string {
	if lineNumber < 1 || lineNumber > t.GetLineCount() {
		return ""
	}

	// Check cache first
	if t.lastVisitedLine.LineNumber == lineNumber {
		return t.lastVisitedLine.Value
	}

	t.lastVisitedLine.LineNumber = lineNumber

	if lineNumber == t.GetLineCount() {
		t.lastVisitedLine.Value = t.GetLineRawContent(lineNumber, 0)
	} else if t.EOLNormalized {
		t.lastVisitedLine.Value = t.GetLineRawContent(lineNumber, t.EOLLength)
	} else {
		rawContent := t.GetLineRawContent(lineNumber, 0)
		// Remove line endings from the end of the line
		t.lastVisitedLine.Value = strings.TrimRight(rawContent, "\r\n")
	}

	return t.lastVisitedLine.Value
}

// GetLineLength 获取指定行的长度
func (t *PieceTreeBase) GetLineLength(lineNumber int) int {
	if lineNumber == t.GetLineCount() {
		startOffset := t.GetOffsetAt(lineNumber, 1)
		return t.GetLength() - startOffset
	}
	return t.GetOffsetAt(lineNumber+1, 1) - t.GetOffsetAt(lineNumber, 1) - t.EOLLength
}

// GetLineCharCode 获取指定行指定列的字符码
func (t *PieceTreeBase) GetLineCharCode(lineNumber, column int) int {
	if lineNumber < 1 || lineNumber > t.lineCnt || column < 1 {
		return 0
	}

	pos := t.NodeAt2(lineNumber, column)
	if pos.Node == nil {
		return 0
	}

	node := pos.Node
	buffer := t.buffers[node.Piece.BufferIndex].Buffer
	startOffset := t.OffsetInBuffer(node.Piece.BufferIndex, node.Piece.Start)
	offset := startOffset + pos.Remainder

	if offset >= len(buffer) {
		return 0
	}

	return int(buffer[offset])
}

// NodeCharCodeAt 获取节点中指定偏移量的字符码
func (t *PieceTreeBase) NodeCharCodeAt(node *TreeNode, offset int) int {
	if node.Piece.LineFeedCnt < 1 {
		return -1
	}
	buffer := t.buffers[node.Piece.BufferIndex]
	newOffset := t.OffsetInBuffer(node.Piece.BufferIndex, node.Piece.Start) + offset
	if newOffset >= len(buffer.Buffer) {
		return -1
	}
	return int(buffer.Buffer[newOffset])
}

// ShouldCheckCRLF 是否应该检查 CRLF
func (t *PieceTreeBase) ShouldCheckCRLF() bool {
	return !(t.EOLNormalized && t.EOL == "\n")
}

// StartWithLF 是否以 LF 开头
func (t *PieceTreeBase) StartWithLF(val interface{}) bool {
	switch v := val.(type) {
	case string:
		if len(v) == 0 {
			return false
		}
		return v[0] == '\n'
	case *TreeNode:
		if v == SENTINEL || v.Piece.LineFeedCnt == 0 {
			return false
		}

		piece := v.Piece
		lineStarts := t.buffers[piece.BufferIndex].LineStarts
		line := piece.Start.Line
		startOffset := lineStarts[line] + piece.Start.Column
		if line == len(lineStarts)-1 {
			// 最后一行，所以这一行末尾没有换行符
			return false
		}
		nextLineOffset := lineStarts[line+1]
		if nextLineOffset > startOffset+1 {
			return false
		}
		return t.buffers[piece.BufferIndex].Buffer[startOffset] == '\n'
	default:
		return false
	}
}

// EndWithCR 是否以 CR 结尾
func (t *PieceTreeBase) EndWithCR(val interface{}) bool {
	switch v := val.(type) {
	case string:
		if len(v) == 0 {
			return false
		}
		return v[len(v)-1] == '\r'
	case *TreeNode:
		if v == SENTINEL || v.Piece.LineFeedCnt == 0 {
			return false
		}

		return t.NodeCharCodeAt(v, v.Piece.Length-1) == '\r'
	default:
		return false
	}
}

// ValidateCRLFWithPrevNode 验证与前一个节点的CRLF
func (t *PieceTreeBase) ValidateCRLFWithPrevNode(node *TreeNode) {
	if !t.ShouldCheckCRLF() {
		return
	}

	if t.StartWithLF(node) && node.Prev() != SENTINEL && t.EndWithCR(t.GetNodeContent(node.Prev())) {
		// 合并 \r\n
		t.FixCRLF(node.Prev(), node)
	}
}

// ValidateCRLFWithNextNode 验证当前节点与后一个节点的 CRLF
func (t *PieceTreeBase) ValidateCRLFWithNextNode(node *TreeNode) {
	if t.ShouldCheckCRLF() && t.EndWithCR(node) {
		nextNode := node.Next()
		if nextNode != nil && nextNode != SENTINEL && t.StartWithLF(nextNode) {
			t.FixCRLF(node, nextNode)
		}
	}
}

// FixCRLF 修复 CRLF
func (t *PieceTreeBase) FixCRLF(prev, next *TreeNode) {
	nodesToDel := make([]*TreeNode, 0)
	// 更新节点
	lineStarts := t.buffers[prev.Piece.BufferIndex].LineStarts
	var newEnd BufferCursor
	if prev.Piece.End.Column == 0 {
		// 表示最后一行以 \r 结尾，而不是 \r\n
		newEnd = BufferCursor{
			Line:   prev.Piece.End.Line - 1,
			Column: lineStarts[prev.Piece.End.Line] - lineStarts[prev.Piece.End.Line-1] - 1,
		}
	} else {
		// \r\n
		newEnd = BufferCursor{
			Line:   prev.Piece.End.Line,
			Column: prev.Piece.End.Column - 1,
		}
	}

	prevNewLength := prev.Piece.Length - 1
	prevNewLFCnt := prev.Piece.LineFeedCnt - 1
	prev.Piece = NewPiece(
		prev.Piece.BufferIndex,
		prev.Piece.Start,
		newEnd,
		prevNewLFCnt,
		prevNewLength,
	)

	UpdateTreeMetadata(t, prev, -1, -1)
	if prev.Piece.Length == 0 {
		nodesToDel = append(nodesToDel, prev)
	}

	// 更新下一个节点
	newStart := BufferCursor{Line: next.Piece.Start.Line + 1, Column: 0}
	newLength := next.Piece.Length - 1
	newLineFeedCnt := t.GetLineFeedCnt(next.Piece.BufferIndex, newStart, next.Piece.End)
	next.Piece = NewPiece(
		next.Piece.BufferIndex,
		newStart,
		next.Piece.End,
		newLineFeedCnt,
		newLength,
	)

	UpdateTreeMetadata(t, next, -1, -1)
	if next.Piece.Length == 0 {
		nodesToDel = append(nodesToDel, next)
	}

	// 创建包含 \r\n 的新片段
	pieces := t.CreateNewPieces("\r\n")
	t.RbInsertRight(prev, pieces[0])

	// 删除空节点
	for i := 0; i < len(nodesToDel); i++ {
		RbDelete(t, nodesToDel[i])
	}
}

// CreateNewPieces 创建新的片段
func (t *PieceTreeBase) CreateNewPieces(text string) []Piece {
	if len(text) > AverageBufferSize {
		newPieces := make([]Piece, 0)
		for len(text) > AverageBufferSize {
			lastChar := text[AverageBufferSize-1]
			var splitText string
			if lastChar == '\r' || (rune(text[AverageBufferSize-1]) >= 0xD800 && rune(text[AverageBufferSize-1]) <= 0xDBFF) {
				// 最后一个字符是 \r 或高代理项，保留它
				splitText = text[:AverageBufferSize-1]
				text = text[AverageBufferSize-1:]
			} else {
				splitText = text[:AverageBufferSize]
				text = text[AverageBufferSize:]
			}

			lineStarts := CreateLineStartsFast(splitText, false)
			newPieces = append(newPieces, NewPiece(
				len(t.buffers), // buffer index
				BufferCursor{Line: 0, Column: 0},
				BufferCursor{
					Line:   len(lineStarts) - 1,
					Column: len(splitText) - lineStarts[len(lineStarts)-1],
				},
				len(lineStarts)-1,
				len(splitText),
			))
			t.buffers = append(t.buffers, NewStringBuffer(splitText, lineStarts))
		}

		lineStarts := CreateLineStartsFast(text, false)
		newPieces = append(newPieces, NewPiece(
			len(t.buffers), // buffer index
			BufferCursor{Line: 0, Column: 0},
			BufferCursor{
				Line:   len(lineStarts) - 1,
				Column: len(text) - lineStarts[len(lineStarts)-1],
			},
			len(lineStarts)-1,
			len(text),
		))
		t.buffers = append(t.buffers, NewStringBuffer(text, lineStarts))

		return newPieces
	}

	startOffset := len(t.buffers[0].Buffer)
	lineStarts := CreateLineStartsFast(text, false)
	start := t.lastChangeBufferPos

	if len(t.buffers[0].LineStarts) > 0 &&
		t.buffers[0].LineStarts[len(t.buffers[0].LineStarts)-1] == startOffset &&
		startOffset != 0 &&
		t.StartWithLF(text) &&
		t.EndWithCR(t.buffers[0].Buffer) {
		t.lastChangeBufferPos = BufferCursor{
			Line:   t.lastChangeBufferPos.Line,
			Column: t.lastChangeBufferPos.Column,
		}
		start = t.lastChangeBufferPos
		for i := 0; i < len(lineStarts); i++ {
			lineStarts[i] = startOffset + 1
		}

		t.buffers[0].LineStarts = append(t.buffers[0].LineStarts, lineStarts[:1]...)
		t.buffers[0].Buffer += "_" + text
		startOffset += 1

	} else {
		if startOffset != 0 {
			for i := 0; i < len(lineStarts); i++ {
				lineStarts[i] += startOffset
			}
		}
		t.buffers[0].LineStarts = append(t.buffers[0].LineStarts, lineStarts[1:]...)
		t.buffers[0].Buffer += text
	}

	endOffset := len(t.buffers[0].Buffer)
	endIndex := len(t.buffers[0].LineStarts) - 1
	endColumn := endOffset - t.buffers[0].LineStarts[endIndex]
	endPos := BufferCursor{Line: endIndex, Column: endColumn}
	newPiece := NewPiece(0, start, endPos, t.GetLineFeedCnt(0, start, endPos), endOffset-startOffset)

	t.lastChangeBufferPos = endPos
	return []Piece{newPiece}
}

// AdjustCarriageReturnFromNext 调整下一个节点的回车符
func (t *PieceTreeBase) AdjustCarriageReturnFromNext(value string, node *TreeNode) bool {
	if t.ShouldCheckCRLF() && t.EndWithCR(value) {
		nextNode := node.Next()
		if t.StartWithLF(nextNode) {
			// 将 \n 向前移动
			value += "\n"

			if nextNode.Piece.Length == 1 {
				RbDelete(t, nextNode)
			} else {
				piece := nextNode.Piece
				newStart := BufferCursor{Line: piece.Start.Line + 1, Column: 0}
				newLength := piece.Length - 1
				newLineFeedCnt := t.GetLineFeedCnt(piece.BufferIndex, newStart, piece.End)
				nextNode.Piece = NewPiece(
					piece.BufferIndex,
					newStart,
					piece.End,
					newLineFeedCnt,
					newLength,
				)

				UpdateTreeMetadata(t, nextNode, -1, -1)
			}
			return true
		}
	}

	return false
}

// InsertContentToNodeLeft 在节点左侧插入内容
func (t *PieceTreeBase) InsertContentToNodeLeft(value string, node *TreeNode) {
	// 我们在节点的开头插入内容
	nodesToDel := make([]*TreeNode, 0)
	if t.ShouldCheckCRLF() && t.EndWithCR(value) && t.StartWithLF(node) {
		// 将 \n 移动到新节点
		piece := node.Piece
		newStart := BufferCursor{Line: piece.Start.Line + 1, Column: 0}
		nPiece := NewPiece(
			piece.BufferIndex,
			newStart,
			piece.End,
			t.GetLineFeedCnt(piece.BufferIndex, newStart, piece.End),
			piece.Length-1,
		)

		node.Piece = nPiece
		value += "\n"
		UpdateTreeMetadata(t, node, -1, -1)

		if node.Piece.Length == 0 {
			nodesToDel = append(nodesToDel, node)
		}
	}

	newPieces := t.CreateNewPieces(value)

	// 从第一个片段开始，依次向左插入
	var newNode *TreeNode = nil
	if len(newPieces) > 0 {
		newNode = t.RbInsertLeft(node, newPieces[0])
		var lastNode *TreeNode = newNode

		for k := 1; k < len(newPieces); k++ {
			lastNode = t.RbInsertRight(lastNode, newPieces[k])
		}
	}

	if newNode != nil {
		t.ValidateCRLFWithPrevNode(newNode)
	}

	// 删除空节点
	for i := 0; i < len(nodesToDel); i++ {
		RbDelete(t, nodesToDel[i])
	}
}

// InsertContentToNodeRight 在节点右侧插入内容
func (t *PieceTreeBase) InsertContentToNodeRight(value string, node *TreeNode) {
	// 我们在节点的右侧插入内容
	if t.AdjustCarriageReturnFromNext(value, node) {
		// 将 \n 移动到新节点
		value += "\n"
	}

	newPieces := t.CreateNewPieces(value)

	if len(newPieces) > 0 {
		newNode := t.RbInsertRight(node, newPieces[0])
		var lastNode *TreeNode = newNode

		for k := 1; k < len(newPieces); k++ {
			lastNode = t.RbInsertRight(lastNode, newPieces[k])
		}

		t.ValidateCRLFWithPrevNode(newNode)
	}
}

// Insert 在指定偏移量处插入内容
func (t *PieceTreeBase) Insert(offset int, value string, eolNormalized bool) {
	// 参数检查和边界处理
	if offset < 0 {
		offset = 0
	}
	if offset > t.length {
		offset = t.length
	}

	// 空字符串不需要插入
	if len(value) == 0 {
		return
	}

	// 更新EOL标志
	t.EOLNormalized = t.EOLNormalized && eolNormalized

	// 重置行缓存
	t.lastVisitedLine.LineNumber = 0
	t.lastVisitedLine.Value = ""

	if t.Root != SENTINEL {
		// 找到插入位置对应的节点
		pos := t.NodeAt(offset)
		if pos.Node == nil {
			return
		}

		node := pos.Node
		remainder := pos.Remainder
		nodeStartOffset := pos.NodeStartOffset
		piece := node.Piece
		bufferIndex := piece.BufferIndex
		insertPosInBuffer := t.PositionInBuffer(node, remainder)

		// 优化：检查是否可以追加到上一次修改的缓冲区
		if node.Piece.BufferIndex == 0 &&
			piece.End.Line == t.lastChangeBufferPos.Line &&
			piece.End.Column == t.lastChangeBufferPos.Column &&
			(nodeStartOffset+piece.Length == offset) &&
			len(value) < AverageBufferSize {
			// 直接追加到已更改的缓冲区
			t.AppendToNode(node, value)
			t.ComputeBufferMetadata()
			return
		}

		if nodeStartOffset == offset {
			// 情况1：在节点开头插入
			t.InsertContentToNodeLeft(value, node)
			t.searchCache.Validate(offset)
		} else if nodeStartOffset+node.Piece.Length > offset {
			// 情况2：在节点中间插入 - 需要分割节点
			nodesToDel := make([]*TreeNode, 0)

			// 创建右侧片段
			newRightPiece := NewPiece(
				piece.BufferIndex,
				insertPosInBuffer,
				piece.End,
				t.GetLineFeedCnt(piece.BufferIndex, insertPosInBuffer, piece.End),
				t.OffsetInBuffer(bufferIndex, piece.End)-t.OffsetInBuffer(bufferIndex, insertPosInBuffer),
			)

			// 处理CRLF边界情况
			if t.ShouldCheckCRLF() && t.EndWithCR(value) {
				headOfRight := t.NodeCharCodeAt(node, remainder)

				if headOfRight == 10 { // \n
					// 如果插入内容以\r结尾，右侧以\n开头，需要特殊处理
					newStart := BufferCursor{Line: newRightPiece.Start.Line + 1, Column: 0}
					newRightPiece = NewPiece(
						newRightPiece.BufferIndex,
						newStart,
						newRightPiece.End,
						t.GetLineFeedCnt(newRightPiece.BufferIndex, newStart, newRightPiece.End),
						newRightPiece.Length-1,
					)

					value += "\n"
				}
			}

			// 处理左侧部分 - 重用当前节点
			if t.ShouldCheckCRLF() && t.StartWithLF(value) {
				// 如果插入内容以\n开头，检查左侧是否以\r结尾
				if remainder > 0 {
					tailOfLeft := t.NodeCharCodeAt(node, remainder-1)
					if tailOfLeft == 13 { // \r
						// 特殊处理CRLF跨节点情况
						previousPos := t.PositionInBuffer(node, remainder-1)
						t.DeleteNodeTail(node, previousPos)
						value = "\r" + value

						if node.Piece.Length == 0 {
							nodesToDel = append(nodesToDel, node)
						}
					} else {
						t.DeleteNodeTail(node, insertPosInBuffer)
					}
				} else {
					t.DeleteNodeTail(node, insertPosInBuffer)
				}
			} else {
				// 正常情况：截断当前节点作为左侧部分
				t.DeleteNodeTail(node, insertPosInBuffer)
			}

			// 创建新片段
			newPieces := t.CreateNewPieces(value)

			// 插入顺序很重要：先插入新内容，再插入右侧片段
			var newNode *TreeNode = nil
			for k := 0; k < len(newPieces); k++ {
				if k == 0 {
					newNode = t.RbInsertRight(node, newPieces[k])
				} else {
					newNode = t.RbInsertRight(newNode, newPieces[k])
				}
			}

			// 插入右侧片段
			if newRightPiece.Length > 0 {
				t.RbInsertRight(newNode, newRightPiece)
			}

			// 删除标记的空节点
			for i := 0; i < len(nodesToDel); i++ {
				RbDelete(t, nodesToDel[i])
			}

			// 验证CRLF连接处
			if newNode != nil {
				t.ValidateCRLFWithPrevNode(newNode)
			}
		} else {
			// 情况3：在节点右侧插入
			t.InsertContentToNodeRight(value, node)
		}
	} else {
		// 空树，插入新节点
		pieces := t.CreateNewPieces(value)
		if len(pieces) > 0 {
			node := t.RbInsertLeft(nil, pieces[0])

			for k := 1; k < len(pieces); k++ {
				node = t.RbInsertRight(node, pieces[k])
			}
		}
	}

	// 更新元数据
	t.ComputeBufferMetadata()
}

// Delete 删除指定范围的内容
func (t *PieceTreeBase) Delete(offset, cnt int) {
	// 参数检查
	if t == nil {
		return
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= t.length {
		return
	}
	if cnt <= 0 {
		return
	}
	// 确保删除范围不超出文本长度
	if offset+cnt > t.length {
		cnt = t.length - offset
	}

	t.lastVisitedLine.LineNumber = 0
	t.lastVisitedLine.Value = ""

	if t.Root == SENTINEL {
		return
	}

	// 记录原始长度用于调试
	originalLength := t.length

	// 特殊处理：如果我们在删除节点边界上的内容，需要找到确切的节点
	// 首先，获取删除范围内的所有节点
	nodesToDelete := make([]*TreeNode, 0)
	var startNode, endNode *TreeNode
	var startRemainder, endRemainder int

	// 找到删除范围内的所有节点
	currentOffset := 0
	t.Iterate(t.Root, func(node *TreeNode) bool {
		if node == SENTINEL {
			return true
		}

		nodeLength := node.Piece.Length
		nodeEndOffset := currentOffset + nodeLength

		// 检查节点是否与删除范围重叠
		if nodeEndOffset > offset && currentOffset < offset+cnt {
			// 节点与删除范围重叠
			if startNode == nil {
				startNode = node
				startRemainder = offset - currentOffset
				if startRemainder < 0 {
					startRemainder = 0
				}
			}

			// 检查是否需要完全删除节点
			if currentOffset >= offset && nodeEndOffset <= offset+cnt {
				// 节点完全在删除范围内
				nodesToDelete = append(nodesToDelete, node)
			} else if currentOffset < offset && nodeEndOffset > offset+cnt {
				// 删除范围在节点内部
				endNode = node
				endRemainder = offset + cnt - currentOffset
			} else if nodeEndOffset > offset+cnt {
				// 节点部分在删除范围内，部分在删除范围外
				endNode = node
				endRemainder = offset + cnt - currentOffset
			}
		}

		currentOffset += nodeLength
		return true
	})

	// 如果没有找到节点，直接返回
	if startNode == nil {
		return
	}

	// 处理起始节点
	if startNode != nil && !contains(nodesToDelete, startNode) {
		// 起始节点需要部分删除
		piece := startNode.Piece
		bufferIndex := piece.BufferIndex

		if startRemainder == 0 {
			// 从节点开头开始删除
			deleteLength := min(cnt, piece.Length)
			newStart := t.PositionInBuffer(startNode, deleteLength)
			newLength := piece.Length - deleteLength
			newLineFeedCnt := t.GetLineFeedCnt(bufferIndex, newStart, piece.End)

			startNode.Piece = NewPiece(
				bufferIndex,
				newStart,
				piece.End,
				newLineFeedCnt,
				newLength,
			)

			// 更新树元数据
			UpdateTreeMetadata(t, startNode, newLength-piece.Length, newLineFeedCnt-piece.LineFeedCnt)

			// 如果节点变为空，删除它
			if newLength == 0 {
				nodesToDelete = append(nodesToDelete, startNode)
			}
		} else if startRemainder > 0 && startRemainder < piece.Length {
			// 删除节点中间部分
			if endNode == startNode {
				// 删除范围在同一个节点内
				leftEnd := t.PositionInBuffer(startNode, startRemainder)
				rightStart := t.PositionInBuffer(startNode, endRemainder)

				// 计算左侧部分
				leftLength := t.OffsetInBuffer(bufferIndex, leftEnd) - t.OffsetInBuffer(bufferIndex, piece.Start)
				leftLineFeedCnt := t.GetLineFeedCnt(bufferIndex, piece.Start, leftEnd)

				// 计算右侧部分
				rightLength := t.OffsetInBuffer(bufferIndex, piece.End) - t.OffsetInBuffer(bufferIndex, rightStart)
				rightLineFeedCnt := t.GetLineFeedCnt(bufferIndex, rightStart, piece.End)

				// 更新当前节点为左侧部分
				startNode.Piece = NewPiece(
					bufferIndex,
					piece.Start,
					leftEnd,
					leftLineFeedCnt,
					leftLength,
				)

				// 更新树元数据
				UpdateTreeMetadata(t, startNode, leftLength-piece.Length, leftLineFeedCnt-piece.LineFeedCnt)

				// 如果左侧部分为空，删除它
				if leftLength == 0 {
					nodesToDelete = append(nodesToDelete, startNode)
				}

				// 如果右侧部分有内容，创建新节点
				if rightLength > 0 {
					rightPiece := NewPiece(
						bufferIndex,
						rightStart,
						piece.End,
						rightLineFeedCnt,
						rightLength,
					)
					t.RbInsertRight(startNode, rightPiece)
				}
			} else {
				// 删除节点结尾部分
				newEnd := t.PositionInBuffer(startNode, startRemainder)
				newLength := startRemainder
				newLineFeedCnt := t.GetLineFeedCnt(bufferIndex, piece.Start, newEnd)

				startNode.Piece = NewPiece(
					bufferIndex,
					piece.Start,
					newEnd,
					newLineFeedCnt,
					newLength,
				)

				// 更新树元数据
				UpdateTreeMetadata(t, startNode, newLength-piece.Length, newLineFeedCnt-piece.LineFeedCnt)

				// 如果节点变为空，删除它
				if newLength == 0 {
					nodesToDelete = append(nodesToDelete, startNode)
				}
			}
		}
	}

	// 处理结束节点
	if endNode != nil && endNode != startNode && !contains(nodesToDelete, endNode) {
		// 结束节点需要部分删除
		piece := endNode.Piece
		bufferIndex := piece.BufferIndex

		if endRemainder > 0 && endRemainder < piece.Length {
			// 删除节点开头部分
			newStart := t.PositionInBuffer(endNode, endRemainder)
			newLength := piece.Length - endRemainder
			newLineFeedCnt := t.GetLineFeedCnt(bufferIndex, newStart, piece.End)

			endNode.Piece = NewPiece(
				bufferIndex,
				newStart,
				piece.End,
				newLineFeedCnt,
				newLength,
			)

			// 更新树元数据
			UpdateTreeMetadata(t, endNode, newLength-piece.Length, newLineFeedCnt-piece.LineFeedCnt)

			// 如果节点变为空，删除它
			if newLength == 0 {
				nodesToDelete = append(nodesToDelete, endNode)
			}
		}
	}

	// 删除标记的节点
	for _, node := range nodesToDelete {
		if node != nil && node != SENTINEL {
			RbDelete(t, node)
		}
	}

	// 验证CRLF连接处
	if startNode != nil && startNode != SENTINEL && startNode.Piece.Length > 0 {
		t.ValidateCRLFWithNextNode(startNode)
	}

	// 更新元数据
	t.ComputeBufferMetadata()

	// 确保删除后的长度正确
	if t.length != originalLength-cnt {
		// 如果长度不正确，强制更新长度
		t.length = originalLength - cnt
	}
}

// contains 检查切片中是否包含指定元素
func contains(slice []*TreeNode, item *TreeNode) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ShrinkNode 缩小节点
func (t *PieceTreeBase) ShrinkNode(node *TreeNode, start, end BufferCursor) {
	piece := node.Piece
	originalStartPos := piece.Start
	originalEndPos := piece.End

	// 旧片段，originalStartPos, start
	oldLength := piece.Length
	oldLFCnt := piece.LineFeedCnt
	newEnd := start
	newLineFeedCnt := t.GetLineFeedCnt(piece.BufferIndex, piece.Start, newEnd)
	newLength := t.OffsetInBuffer(piece.BufferIndex, start) - t.OffsetInBuffer(piece.BufferIndex, originalStartPos)

	node.Piece = NewPiece(
		piece.BufferIndex,
		piece.Start,
		newEnd,
		newLineFeedCnt,
		newLength,
	)

	UpdateTreeMetadata(t, node, newLength-oldLength, newLineFeedCnt-oldLFCnt)

	// 新的右侧片段，end, originalEndPos
	newPiece := NewPiece(
		piece.BufferIndex,
		end,
		originalEndPos,
		t.GetLineFeedCnt(piece.BufferIndex, end, originalEndPos),
		t.OffsetInBuffer(piece.BufferIndex, originalEndPos)-t.OffsetInBuffer(piece.BufferIndex, end),
	)

	newNode := t.RbInsertRight(node, newPiece)
	t.ValidateCRLFWithPrevNode(newNode)
}

// GetOffsetAt 根据行号和列号获取偏移量
func (t *PieceTreeBase) GetOffsetAt(lineNumber, column int) int {
	leftLen := 0 // inorder

	x := t.Root

	for x != SENTINEL {
		if x.Left != SENTINEL && x.LFLeft+1 >= lineNumber {
			x = x.Left
		} else if x.LFLeft+x.Piece.LineFeedCnt+1 >= lineNumber {
			leftLen += x.SizeLeft
			// lineNumber >= 2
			accumulatedValInCurrentIndex := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-2)
			return leftLen + accumulatedValInCurrentIndex + column - 1
		} else {
			lineNumber -= x.LFLeft + x.Piece.LineFeedCnt
			leftLen += x.SizeLeft + x.Piece.Length
			x = x.Right
		}
	}

	return leftLen
}

// GetValueInRange 获取指定范围内的值
func (t *PieceTreeBase) GetValueInRange(startLineNumber, startColumn, endLineNumber, endColumn int, eol string) string {
	// 如果起始位置和结束位置相同，返回空字符串
	if startLineNumber == endLineNumber && startColumn == endColumn {
		return ""
	}

	// 获取起始位置和结束位置对应的节点
	startPosition := t.NodeAt2(startLineNumber, startColumn)
	endPosition := t.NodeAt2(endLineNumber, endColumn)

	// 获取范围内的值
	value := t.GetValueInRange2(startPosition, endPosition)

	// 如果指定了换行符，进行替换
	if eol != "" {
		if eol != t.EOL || !t.EOLNormalized {
			// 使用正则表达式替换所有换行符
			re := regexp.MustCompile(`\r\n|\r|\n`)
			return re.ReplaceAllString(value, eol)
		}

		if eol == t.GetEOL() && t.EOLNormalized {
			return value
		}

		// 使用正则表达式替换所有换行符
		re := regexp.MustCompile(`\r\n|\r|\n`)
		return re.ReplaceAllString(value, eol)
	}

	return value
}

// AppendToNode 向节点追加内容
func (t *PieceTreeBase) AppendToNode(node *TreeNode, value string) {
	if t.AdjustCarriageReturnFromNext(value, node) {
		value += "\n"
	}

	hitCRLF := t.ShouldCheckCRLF() && t.StartWithLF(value) && t.EndWithCR(node)
	startOffset := len(t.buffers[0].Buffer)
	t.buffers[0].Buffer += value
	lineStarts := CreateLineStartsFast(value, false)
	for i := 0; i < len(lineStarts); i++ {
		lineStarts[i] += startOffset
	}

	if hitCRLF {
		prevStartOffset := t.buffers[0].LineStarts[len(t.buffers[0].LineStarts)-2]
		t.buffers[0].LineStarts = t.buffers[0].LineStarts[:len(t.buffers[0].LineStarts)-1]
		// lastChangeBufferPos 已经错误
		t.lastChangeBufferPos = BufferCursor{Line: t.lastChangeBufferPos.Line - 1, Column: startOffset - prevStartOffset}
	}

	t.buffers[0].LineStarts = append(t.buffers[0].LineStarts, lineStarts[1:]...)
	endIndex := len(t.buffers[0].LineStarts) - 1
	endColumn := len(t.buffers[0].Buffer) - t.buffers[0].LineStarts[endIndex]
	newEnd := BufferCursor{Line: endIndex, Column: endColumn}
	newLength := node.Piece.Length + len(value)
	oldLineFeedCnt := node.Piece.LineFeedCnt
	newLineFeedCnt := t.GetLineFeedCnt(0, node.Piece.Start, newEnd)
	lfDelta := newLineFeedCnt - oldLineFeedCnt

	node.Piece = NewPiece(
		node.Piece.BufferIndex,
		node.Piece.Start,
		newEnd,
		newLineFeedCnt,
		newLength,
	)

	t.lastChangeBufferPos = newEnd
	UpdateTreeMetadata(t, node, len(value), lfDelta)
}

// FindLastNode 查找树中的最后一个节点
func (t *PieceTreeBase) FindLastNode() *TreeNode {
	if t.Root == SENTINEL {
		return SENTINEL
	}

	// 从根节点开始，一直向右查找到最右边的节点
	current := t.Root
	for current.Right != SENTINEL {
		current = current.Right
	}

	// 如果最右节点有子节点，找到其最右子节点
	if current.Piece.LineFeedCnt > 0 {
		// 找到最后一行的最后一个节点
		return current
	}

	// 查找下一个节点，直到找到最后一个
	next := current.Next()
	for next != SENTINEL {
		current = next
		next = next.Next()
	}

	return current
}

// GetLinesContent 获取所有行的内容
func (t *PieceTreeBase) GetLinesContent() []string {
	content := t.GetContentOfSubTree(t.Root)
	if content == "" {
		return []string{""}
	}
	return strings.Split(content, t.EOL)
}

// GetLinesRawContent 获取所有行的原始内容
func (t *PieceTreeBase) GetLinesRawContent() string {
	return t.GetContentOfSubTree(t.Root)
}
