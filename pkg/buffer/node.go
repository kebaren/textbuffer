package buffer

// NodeColor 节点颜色
type NodeColor int

const (
	// Black 黑色
	Black NodeColor = 0
	// Red 红色
	Red NodeColor = 1
)

// BufferCursor 缓冲区游标
type BufferCursor struct {
	// Line 行
	Line int
	// Column 列
	Column int
}

// Piece 片段
type Piece struct {
	// BufferIndex 缓冲区索引
	BufferIndex int
	// Start 开始位置
	Start BufferCursor
	// End 结束位置
	End BufferCursor
	// Length 长度
	Length int
	// LineFeedCnt 换行符计数
	LineFeedCnt int
}

// NewPiece 创建一个新的片段
func NewPiece(bufferIndex int, start, end BufferCursor, lineFeedCnt, length int) Piece {
	return Piece{
		BufferIndex: bufferIndex,
		Start:       start,
		End:         end,
		LineFeedCnt: lineFeedCnt,
		Length:      length,
	}
}

// TreeNode 树节点
type TreeNode struct {
	// Color 颜色
	Color NodeColor
	// SizeLeft 左侧大小
	SizeLeft int
	// LFLeft 左侧换行符计数
	LFLeft int
	// Left 左子节点
	Left *TreeNode
	// Right 右子节点
	Right *TreeNode
	// Parent 父节点
	Parent *TreeNode
	// Piece 片段
	Piece Piece
}

// NewTreeNode 创建一个新的树节点
func NewTreeNode(piece Piece, color NodeColor) *TreeNode {
	return &TreeNode{
		Color:    color,
		SizeLeft: 0,
		LFLeft:   0,
		Left:     nil,
		Right:    nil,
		Parent:   nil,
		Piece:    piece,
	}
}

// SENTINEL 哨兵节点
var SENTINEL = &TreeNode{
	Color:    Black,
	SizeLeft: 0,
	LFLeft:   0,
	Left:     nil,
	Right:    nil,
	Parent:   nil,
	Piece: Piece{
		BufferIndex: 0,
		Start:       BufferCursor{Line: 0, Column: 0},
		End:         BufferCursor{Line: 0, Column: 0},
		LineFeedCnt: 0,
		Length:      0,
	},
}

// NodePosition 节点位置
type NodePosition struct {
	// Node 节点
	Node *TreeNode
	// Remainder 剩余
	Remainder int
	// NodeStartOffset 节点开始偏移量
	NodeStartOffset int
}

// Next 获取下一个节点
func (n *TreeNode) Next() *TreeNode {
	if n.Right != SENTINEL {
		return Leftest(n.Right)
	}

	var p *TreeNode = n.Parent
	for p != SENTINEL && n == p.Right {
		n = p
		p = p.Parent
	}

	return p
}

// Prev 获取上一个节点
func (n *TreeNode) Prev() *TreeNode {
	if n.Left != SENTINEL {
		return Righttest(n.Left)
	}

	var p *TreeNode = n.Parent
	for p != SENTINEL && n == p.Left {
		n = p
		p = p.Parent
	}

	return p
}

// Leftest 获取最左侧节点
func Leftest(node *TreeNode) *TreeNode {
	if node == SENTINEL {
		return SENTINEL
	}

	for node.Left != SENTINEL {
		node = node.Left
	}

	return node
}

// Righttest 获取最右侧节点
func Righttest(node *TreeNode) *TreeNode {
	if node == SENTINEL {
		return SENTINEL
	}

	for node.Right != SENTINEL {
		node = node.Right
	}

	return node
}

// FixInsert 修复插入
func FixInsert(tree *PieceTreeBase, x *TreeNode) {
	RecomputeTreeMetadata(tree, x)

	for x != tree.Root && x.Parent.Color == Red {
		if x.Parent == x.Parent.Parent.Left {
			y := x.Parent.Parent.Right

			if y.Color == Red {
				x.Parent.Color = Black
				y.Color = Black
				x.Parent.Parent.Color = Red
				x = x.Parent.Parent
			} else {
				if x == x.Parent.Right {
					x = x.Parent
					LeftRotate(tree, x)
				}

				x.Parent.Color = Black
				x.Parent.Parent.Color = Red
				RightRotate(tree, x.Parent.Parent)
			}
		} else {
			y := x.Parent.Parent.Left

			if y.Color == Red {
				x.Parent.Color = Black
				y.Color = Black
				x.Parent.Parent.Color = Red
				x = x.Parent.Parent
			} else {
				if x == x.Parent.Left {
					x = x.Parent
					RightRotate(tree, x)
				}

				x.Parent.Color = Black
				x.Parent.Parent.Color = Red
				LeftRotate(tree, x.Parent.Parent)
			}
		}
	}

	tree.Root.Color = Black
}

// LeftRotate 左旋转
func LeftRotate(tree *PieceTreeBase, x *TreeNode) {
	y := x.Right
	x.Right = y.Left

	if y.Left != SENTINEL {
		y.Left.Parent = x
	}

	y.Parent = x.Parent

	if x.Parent == SENTINEL {
		tree.Root = y
	} else if x == x.Parent.Left {
		x.Parent.Left = y
	} else {
		x.Parent.Right = y
	}

	y.Left = x
	x.Parent = y

	y.SizeLeft += x.SizeLeft
	y.LFLeft += x.LFLeft
}

// RightRotate 右旋转
func RightRotate(tree *PieceTreeBase, y *TreeNode) {
	x := y.Left
	y.Left = x.Right

	if x.Right != SENTINEL {
		x.Right.Parent = y
	}

	x.Parent = y.Parent

	if y.Parent == SENTINEL {
		tree.Root = x
	} else if y == y.Parent.Right {
		y.Parent.Right = x
	} else {
		y.Parent.Left = x
	}

	x.Right = y
	y.Parent = x

	y.SizeLeft -= x.SizeLeft
	y.LFLeft -= x.LFLeft
}

// Detach 分离节点
func (n *TreeNode) Detach() {
	n.Parent = nil
	n.Left = nil
	n.Right = nil
}
