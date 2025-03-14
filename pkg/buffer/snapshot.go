package buffer

// ITextSnapshot 文本快照接口
type ITextSnapshot interface {
	// Read 读取快照内容
	Read() string
}

// PieceTreeSnapshot 片段树快照
type PieceTreeSnapshot struct {
	// pieces 片段数组
	pieces []*TreeNode
	// index 当前索引
	index int
	// tree 片段树
	tree *PieceTreeBase
	// BOM 字节顺序标记
	BOM string
}

// NewPieceTreeSnapshot 创建一个新的片段树快照
func NewPieceTreeSnapshot(tree *PieceTreeBase, BOM string) *PieceTreeSnapshot {
	s := &PieceTreeSnapshot{
		pieces: make([]*TreeNode, 0),
		index:  0,
		tree:   tree,
		BOM:    BOM,
	}

	// 如果根节点不是哨兵，则从树中填充片段
	if tree.Root != SENTINEL {
		tree.Iterate(tree.Root, func(node *TreeNode) bool {
			s.pieces = append(s.pieces, node)
			return true
		})
	}

	return s
}

// Read 读取快照内容
func (s *PieceTreeSnapshot) Read() string {
	if len(s.pieces) == 0 {
		return s.BOM
	}

	result := s.BOM
	for _, piece := range s.pieces {
		result += s.tree.GetPieceContent(piece.Piece)
	}

	return result
}
