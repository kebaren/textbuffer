package buffer

import (
	"regexp"
	"textbuffer/buffer/common"
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
	for i, length := 0, len(chunks); i < length; i++ {
		if len(chunks[i].Buffer) > 0 {
			if chunks[i].LineStarts == nil || len(chunks[i].LineStarts) == 0 {
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
	if node == SENTINEL {
		return ""
	}
	buffer := t.buffers[node.Piece.BufferIndex]
	piece := node.Piece
	startOffset := t.OffsetInBuffer(piece.BufferIndex, piece.Start)
	endOffset := t.OffsetInBuffer(piece.BufferIndex, piece.End)
	currentContent := buffer.Buffer[startOffset:endOffset]
	return currentContent
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
		if x.SizeLeft > offset {
			x = x.Left
		} else if x.SizeLeft+x.Piece.Length >= offset {
			nodeStartOffset += x.SizeLeft
			ret := NodePosition{
				Node:            x,
				Remainder:       offset - x.SizeLeft,
				NodeStartOffset: nodeStartOffset,
			}
			t.searchCache.Set(CacheEntry{
				Node:            x,
				NodeStartOffset: nodeStartOffset,
			})
			return ret
		} else {
			offset -= x.SizeLeft + x.Piece.Length
			nodeStartOffset += x.SizeLeft + x.Piece.Length
			x = x.Right
		}
	}

	return NodePosition{}
}

// NodeAt2 根据行号和列号获取节点位置
func (t *PieceTreeBase) NodeAt2(lineNumber, column int) NodePosition {
	x := t.Root
	nodeStartOffset := 0

	for x != SENTINEL {
		if x.Left != SENTINEL && x.LFLeft >= lineNumber-1 {
			x = x.Left
		} else if x.LFLeft+x.Piece.LineFeedCnt > lineNumber-1 {
			// 在当前节点内的某一行
			prevAccumulatedValue := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-2)
			accumulatedValue := t.GetAccumulatedValue(x, lineNumber-x.LFLeft-1)
			nodeStartOffset += x.SizeLeft

			return NodePosition{
				Node:            x,
				Remainder:       min(prevAccumulatedValue+column-1, accumulatedValue),
				NodeStartOffset: nodeStartOffset,
			}
		} else if x.LFLeft+x.Piece.LineFeedCnt == lineNumber-1 {
			// 在当前节点的最后一行
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

	// 按顺序搜索，找到包含 position.column 的节点
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
		// 空树，直接设置为根节点
		t.Root = z
		z.Color = Black
		return z
	}

	if node == nil {
		// 如果node为nil，插入到树的最左侧
		if t.Root != SENTINEL {
			leftmost := t.Root
			for leftmost.Left != SENTINEL {
				leftmost = leftmost.Left
			}
			leftmost.Left = z
			z.Parent = leftmost
		} else {
			// 树为空，设置为根节点
			t.Root = z
			z.Color = Black
		}
	} else if node == SENTINEL {
		// 如果node是SENTINEL，插入到树的最左侧
		if t.Root != SENTINEL {
			leftmost := t.Root
			for leftmost.Left != SENTINEL {
				leftmost = leftmost.Left
			}
			leftmost.Left = z
			z.Parent = leftmost
		} else {
			// 树为空，设置为根节点
			t.Root = z
			z.Color = Black
		}
	} else if node.Right == SENTINEL {
		// 如果node的右子节点为SENTINEL，直接插入
		node.Right = z
		z.Parent = node
	} else {
		// 否则，插入到node右子树的最左侧
		nextNode := node.Right
		for nextNode.Left != SENTINEL {
			nextNode = nextNode.Left
		}
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
		// 空树，直接设置为根节点
		t.Root = z
		z.Color = Black
		return z
	}

	if node == nil || node == SENTINEL {
		// 如果node为nil或SENTINEL，插入到树的最左侧
		if t.Root != SENTINEL {
			leftmost := t.Root
			for leftmost.Left != SENTINEL {
				leftmost = leftmost.Left
			}
			leftmost.Left = z
			z.Parent = leftmost
		} else {
			// 树为空，设置为根节点
			t.Root = z
			z.Color = Black
		}
	} else if node.Left == SENTINEL {
		// 如果node的左子节点为SENTINEL，直接插入
		node.Left = z
		z.Parent = node
	} else {
		// 否则，插入到node左子树的最右侧
		prevNode := node.Left
		for prevNode.Right != SENTINEL {
			prevNode = prevNode.Right
		}
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

	// 如果end.column为0，说明end正好在行首，不需要特殊处理CRLF
	if end.Column == 0 {
		return end.Line - start.Line
	}

	startOffset := t.OffsetInBuffer(bufferIndex, start)
	endOffset := t.OffsetInBuffer(bufferIndex, end)

	// 确保偏移量有效
	if startOffset >= endOffset {
		return 0
	}

	buffer := t.buffers[bufferIndex].Buffer

	// 确保不越界
	if endOffset > len(buffer) {
		endOffset = len(buffer)
	}

	text := buffer[startOffset:endOffset]

	count := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			count++
		} else if text[i] == '\r' {
			// 处理CRLF序列
			if i+1 < len(text) && text[i+1] == '\n' {
				count++
				i++ // 跳过下一个字符，因为已经计算了CRLF
			} else {
				// 单独的\r也算一个换行
				count++
			}
		}
	}

	return count
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
	} else {
		return lineStarts[expectedLineStartIndex] - lineStarts[piece.Start.Line] - piece.Start.Column
	}
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
		// 我们正在检查此节点的末尾，因此需要进行 CRLF 检查
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

// GetLinesContent 获取所有行的内容
func (t *PieceTreeBase) GetLinesContent() []string {
	result := make([]string, 0, t.lineCnt)
	if t.Root == SENTINEL {
		return result
	}

	// 添加第一行
	result = append(result, "")

	t.Iterate(t.Root, func(node *TreeNode) bool {
		if node == SENTINEL {
			return true
		}

		piece := node.Piece
		buffer := t.buffers[piece.BufferIndex].Buffer
		startOffset := t.OffsetInBuffer(piece.BufferIndex, piece.Start)
		endOffset := t.OffsetInBuffer(piece.BufferIndex, piece.End)

		// 处理此片段中的每个字符
		for i := startOffset; i < endOffset; i++ {
			ch := buffer[i]
			if ch == '\r' {
				if i+1 < endOffset && buffer[i+1] == '\n' {
					// 跳过 \r，\n 将在下一次迭代中处理
					continue
				}
				// 单独的 \r 作为换行符
				result = append(result, "")
			} else if ch == '\n' {
				// \n 作为换行符
				result = append(result, "")
			} else {
				// 普通字符，添加到当前行
				lastIdx := len(result) - 1
				result[lastIdx] += string(ch)
			}
		}

		return true
	})

	return result
}

// GetLineContent 获取指定行的内容
func (t *PieceTreeBase) GetLineContent(lineNumber int) string {
	if lineNumber < 1 || lineNumber > t.lineCnt {
		return ""
	}

	// 检查缓存
	if t.lastVisitedLine.LineNumber == lineNumber {
		return t.lastVisitedLine.Value
	}

	// 获取行内容
	content := ""
	if lineNumber == t.lineCnt {
		// 最后一行
		pos := t.NodeAt2(lineNumber, 1)
		if pos.Node == nil {
			return ""
		}

		node := pos.Node
		buffer := t.buffers[node.Piece.BufferIndex].Buffer
		startOffset := t.OffsetInBuffer(node.Piece.BufferIndex, node.Piece.Start)
		content = buffer[startOffset+pos.Remainder:]

		// 继续遍历后续节点
		node = node.Next()
		for node != SENTINEL {
			content += t.GetNodeContent(node)
			node = node.Next()
		}
	} else {
		// 非最后一行
		startPos := t.NodeAt2(lineNumber, 1)
		endPos := t.NodeAt2(lineNumber+1, 1)
		if startPos.Node == nil || endPos.Node == nil {
			return ""
		}

		// 获取行内容
		content = t.GetValueInRange2(startPos, endPos)
		// 移除行尾的换行符
		if len(content) > 0 {
			if content[len(content)-1] == '\n' {
				content = content[:len(content)-1]
				if len(content) > 0 && content[len(content)-1] == '\r' {
					content = content[:len(content)-1]
				}
			} else if content[len(content)-1] == '\r' {
				content = content[:len(content)-1]
			}
		}
	}

	// 更新缓存
	t.lastVisitedLine.LineNumber = lineNumber
	t.lastVisitedLine.Value = content

	return content
}

// GetLineLength 获取指定行的长度
func (t *PieceTreeBase) GetLineLength(lineNumber int) int {
	if lineNumber < 1 || lineNumber > t.lineCnt {
		return 0
	}

	if lineNumber == t.lineCnt {
		// 最后一行，没有换行符
		return len(t.GetLineContent(lineNumber))
	}

	// 非最后一行，需要考虑换行符
	startPos := t.NodeAt2(lineNumber, 1)
	endPos := t.NodeAt2(lineNumber+1, 1)
	if startPos.Node == nil || endPos.Node == nil {
		return 0
	}

	// 获取行内容（包括换行符）
	content := t.GetValueInRange2(startPos, endPos)

	// 计算长度（不包括换行符）
	length := len(content)
	if length > 0 {
		if content[length-1] == '\n' {
			length--
			if length > 0 && content[length-1] == '\r' {
				length--
			}
		} else if content[length-1] == '\r' {
			length--
		}
	}

	return length
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
		if t.StartWithLF(nextNode) {
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
	if len(text) == 0 {
		return []Piece{}
	}

	// 计算行起始位置
	lineStarts := CreateLineStartsFast(text, false)
	if len(lineStarts) == 0 {
		lineStarts = []int{0}
	}

	start := t.lastChangeBufferPos
	if len(t.buffers[0].LineStarts) > 0 &&
		len(t.buffers[0].Buffer) > 0 &&
		len(t.buffers[0].LineStarts) > int(start.Line) &&
		t.buffers[0].LineStarts[start.Line]+start.Column == len(t.buffers[0].Buffer) &&
		len(text) > 0 && text[0] == '\n' &&
		t.EndWithCR(t.buffers[0].Buffer) {
		// 特殊处理CRLF情况
		t.lastChangeBufferPos = BufferCursor{Line: t.lastChangeBufferPos.Line, Column: t.lastChangeBufferPos.Column + 1}
		start = t.lastChangeBufferPos

		startOffset := 0
		if len(lineStarts) > 0 {
			startOffset = lineStarts[0]
		}

		for i := 0; i < len(lineStarts); i++ {
			lineStarts[i] += startOffset + 1
		}

		t.buffers[0].LineStarts = append(t.buffers[0].LineStarts, lineStarts[1:]...)
		t.buffers[0].Buffer += "_" + text
		startOffset += 1
	} else {
		// 正常情况
		startOffset := 0
		if len(lineStarts) > 0 {
			startOffset = lineStarts[0]
		}

		if startOffset != 0 {
			for i := 0; i < len(lineStarts); i++ {
				lineStarts[i] += startOffset
			}
		}

		// 确保LineStarts至少有一个元素
		if len(t.buffers[0].LineStarts) == 0 {
			t.buffers[0].LineStarts = append(t.buffers[0].LineStarts, 0)
		}

		t.buffers[0].LineStarts = append(t.buffers[0].LineStarts, lineStarts[1:]...)
		t.buffers[0].Buffer += text
	}

	endOffset := len(t.buffers[0].Buffer)
	endIndex := len(t.buffers[0].LineStarts) - 1

	// 确保endIndex不越界
	if endIndex < 0 {
		endIndex = 0
	}

	endColumn := 0
	if endIndex < len(t.buffers[0].LineStarts) {
		endColumn = endOffset - t.buffers[0].LineStarts[endIndex]
	}

	endPos := BufferCursor{Line: endIndex, Column: endColumn}

	// 计算换行符数量
	lineFeedCnt := t.GetLineFeedCnt(0, start, endPos)

	newPiece := NewPiece(
		0,
		start,
		endPos,
		lineFeedCnt,
		endOffset-t.OffsetInBuffer(0, start),
	)
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
	if len(newPieces) == 0 {
		return
	}

	// 从第一个片段开始，依次向左插入
	var newNode *TreeNode
	for k := 0; k < len(newPieces); k++ {
		if newNode == nil {
			newNode = t.RbInsertLeft(node, newPieces[k])
		} else {
			newNode = t.RbInsertRight(newNode, newPieces[k])
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
	newNode := t.RbInsertRight(node, newPieces[0])
	tmpNode := newNode

	for k := 1; k < len(newPieces); k++ {
		tmpNode = t.RbInsertRight(tmpNode, newPieces[k])
	}

	t.ValidateCRLFWithPrevNode(newNode)
}

// Insert 在指定偏移量处插入内容
func (t *PieceTreeBase) Insert(offset int, value string, eolNormalized bool) {
	t.EOLNormalized = t.EOLNormalized && eolNormalized
	t.lastVisitedLine.LineNumber = 0
	t.lastVisitedLine.Value = ""

	if t.Root != SENTINEL {
		pos := t.NodeAt(offset)
		node := pos.Node
		remainder := pos.Remainder
		nodeStartOffset := pos.NodeStartOffset
		piece := node.Piece
		bufferIndex := piece.BufferIndex
		insertPosInBuffer := t.PositionInBuffer(node, remainder)

		// 检查是否可以追加到上一次修改的缓冲区
		if node.Piece.BufferIndex == 0 &&
			piece.End.Line == t.lastChangeBufferPos.Line &&
			piece.End.Column == t.lastChangeBufferPos.Column &&
			(nodeStartOffset+piece.Length == offset) &&
			len(value) < AverageBufferSize {
			// 追加到已更改的缓冲区
			t.AppendToNode(node, value)
			t.ComputeBufferMetadata()
			return
		}

		if nodeStartOffset == offset {
			// 在节点开头插入
			t.InsertContentToNodeLeft(value, node)
			t.searchCache.Validate(offset)
		} else if nodeStartOffset+node.Piece.Length > offset {
			// 在节点中间插入
			nodesToDel := make([]*TreeNode, 0)
			newRightPiece := NewPiece(
				piece.BufferIndex,
				insertPosInBuffer,
				piece.End,
				t.GetLineFeedCnt(piece.BufferIndex, insertPosInBuffer, piece.End),
				t.OffsetInBuffer(bufferIndex, piece.End)-t.OffsetInBuffer(bufferIndex, insertPosInBuffer),
			)

			// 检查 CRLF
			if t.ShouldCheckCRLF() && t.EndWithCR(value) {
				headOfRight := t.NodeCharCodeAt(node, remainder)

				if headOfRight == 10 { // \n
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

			// 重用节点作为插入点之前的内容
			if t.ShouldCheckCRLF() && t.StartWithLF(value) {
				tailOfLeft := t.NodeCharCodeAt(node, remainder-1)
				if tailOfLeft == 13 { // \r
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

			// 创建新片段
			newPieces := t.CreateNewPieces(value)

			// 插入右侧片段
			tmpNode := node
			if newRightPiece.Length > 0 {
				tmpNode = t.RbInsertRight(node, newRightPiece)
			}

			// 插入新片段
			for k := 0; k < len(newPieces); k++ {
				tmpNode = t.RbInsertRight(tmpNode, newPieces[k])
			}

			// 删除标记的节点
			for i := 0; i < len(nodesToDel); i++ {
				RbDelete(t, nodesToDel[i])
			}

			// 验证 CRLF
			t.ValidateCRLFWithPrevNode(tmpNode)
		} else {
			// 在节点右侧插入
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
	t.lastVisitedLine.LineNumber = 0
	t.lastVisitedLine.Value = ""

	if cnt <= 0 || t.Root == SENTINEL {
		return
	}

	startPosition := t.NodeAt(offset)
	endPosition := t.NodeAt(offset + cnt)
	startNode := startPosition.Node
	endNode := endPosition.Node

	if startNode == endNode {
		// 删除在同一个节点内
		startSplitPosInBuffer := t.PositionInBuffer(startNode, startPosition.Remainder)
		endSplitPosInBuffer := t.PositionInBuffer(startNode, endPosition.Remainder)

		if startPosition.NodeStartOffset == offset {
			// 从节点开头删除
			if cnt == startNode.Piece.Length {
				// 删除整个节点
				next := startNode.Next()
				RbDelete(t, startNode)
				t.ValidateCRLFWithPrevNode(next)
				t.ComputeBufferMetadata()
				return
			}
			// 删除节点头部
			t.DeleteNodeHead(startNode, endSplitPosInBuffer)
			t.searchCache.Validate(offset)
			t.ValidateCRLFWithPrevNode(startNode)
			t.ComputeBufferMetadata()
			return
		}

		if startPosition.NodeStartOffset+startNode.Piece.Length == offset+cnt {
			// 删除节点尾部
			t.DeleteNodeTail(startNode, startSplitPosInBuffer)
			t.ValidateCRLFWithNextNode(startNode)
			t.ComputeBufferMetadata()
			return
		}

		// 删除节点中间部分，需要拆分节点
		t.ShrinkNode(startNode, startSplitPosInBuffer, endSplitPosInBuffer)
		t.ComputeBufferMetadata()
		return
	}

	// 删除跨越多个节点
	nodesToDel := make([]*TreeNode, 0)

	// 处理起始节点
	startSplitPosInBuffer := t.PositionInBuffer(startNode, startPosition.Remainder)
	t.DeleteNodeTail(startNode, startSplitPosInBuffer)
	t.searchCache.Validate(offset)
	if startNode.Piece.Length == 0 {
		nodesToDel = append(nodesToDel, startNode)
	}

	// 处理结束节点
	endSplitPosInBuffer := t.PositionInBuffer(endNode, endPosition.Remainder)
	t.DeleteNodeHead(endNode, endSplitPosInBuffer)
	if endNode.Piece.Length == 0 {
		nodesToDel = append(nodesToDel, endNode)
	}

	// 删除中间的节点
	secondNode := startNode.Next()
	for node := secondNode; node != SENTINEL && node != endNode; node = node.Next() {
		nodesToDel = append(nodesToDel, node)
	}

	// 获取前一个有效节点
	prev := startNode
	if startNode.Piece.Length == 0 {
		prev = startNode.Prev()
	}

	// 删除标记的节点
	for i := 0; i < len(nodesToDel); i++ {
		RbDelete(t, nodesToDel[i])
	}

	// 验证 CRLF
	t.ValidateCRLFWithNextNode(prev)
	t.ComputeBufferMetadata()
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
