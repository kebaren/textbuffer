package buffer

// CalculateSize 计算节点及其右子树的大小
func CalculateSize(node *TreeNode) int {
	if node == nil || node == SENTINEL {
		return 0
	}

	return node.SizeLeft + node.Piece.Length + CalculateSize(node.Right)
}

// CalculateLF 计算节点及其右子树的换行符数量
func CalculateLF(node *TreeNode) int {
	if node == nil || node == SENTINEL {
		return 0
	}

	return node.LFLeft + node.Piece.LineFeedCnt + CalculateLF(node.Right)
}

// ResetSentinel 重置哨兵节点
func ResetSentinel() {
	SENTINEL.Parent = SENTINEL
}

// UpdateTreeMetadata 更新树的元数据
func UpdateTreeMetadata(tree *PieceTreeBase, x *TreeNode, delta, lineFeedCntDelta int) {
	// 节点长度变化或换行符数量变化
	for x != tree.Root && x != SENTINEL {
		if x.Parent.Left == x {
			x.Parent.SizeLeft += delta
			x.Parent.LFLeft += lineFeedCntDelta
		}

		x = x.Parent
	}
}

// RecomputeTreeMetadata 重新计算树的元数据
func RecomputeTreeMetadata(tree *PieceTreeBase, x *TreeNode) {
	if x == tree.Root {
		return
	}

	// 向上遍历直到左子树发生变化的节点
	for x != tree.Root && x == x.Parent.Right {
		x = x.Parent
	}

	if x == tree.Root {
		// 表示我们在末尾添加了一个节点（中序）
		return
	}

	// x 是左子树发生变化的节点的父节点
	x = x.Parent

	delta := CalculateSize(x.Left) - x.SizeLeft
	lfDelta := CalculateLF(x.Left) - x.LFLeft
	x.SizeLeft += delta
	x.LFLeft += lfDelta

	// 向上遍历直到根节点，时间复杂度 O(logN)
	for x != tree.Root && (delta != 0 || lfDelta != 0) {
		if x.Parent.Left == x {
			x.Parent.SizeLeft += delta
			x.Parent.LFLeft += lfDelta
		}

		x = x.Parent
	}
}

// RbDelete 删除节点
func RbDelete(tree *PieceTreeBase, z *TreeNode) {
	var x, y *TreeNode

	if z.Left == SENTINEL {
		y = z
		x = y.Right
	} else if z.Right == SENTINEL {
		y = z
		x = y.Left
	} else {
		y = Leftest(z.Right)
		x = y.Right
	}

	if y == tree.Root {
		tree.Root = x

		// 如果 x 为空，我们正在删除唯一的节点
		x.Color = Black
		z.Detach()
		ResetSentinel()
		tree.Root.Parent = SENTINEL

		return
	}

	yWasRed := (y.Color == Red)

	if y == y.Parent.Left {
		y.Parent.Left = x
	} else {
		y.Parent.Right = x
	}

	if y == z {
		x.Parent = y.Parent
		RecomputeTreeMetadata(tree, x)
	} else {
		if y.Parent == z {
			x.Parent = y
		} else {
			x.Parent = y.Parent
		}

		// 当我们对 x 的层次结构进行更改时，首先更新子树的 SizeLeft
		RecomputeTreeMetadata(tree, x)

		y.Left = z.Left
		y.Right = z.Right
		y.Parent = z.Parent
		y.Color = z.Color

		if z == tree.Root {
			tree.Root = y
		} else {
			if z == z.Parent.Left {
				z.Parent.Left = y
			} else {
				z.Parent.Right = y
			}
		}

		if y.Left != SENTINEL {
			y.Left.Parent = y
		}
		if y.Right != SENTINEL {
			y.Right.Parent = y
		}
		// 更新元数据
		// 我们用 y 替换 z，所以在这个子树中，长度变化是 z.Piece.Length
		y.SizeLeft = z.SizeLeft
		y.LFLeft = z.LFLeft
		RecomputeTreeMetadata(tree, y)
	}

	z.Detach()

	if x.Parent.Left == x {
		newSizeLeft := CalculateSize(x)
		newLFLeft := CalculateLF(x)
		if newSizeLeft != x.Parent.SizeLeft || newLFLeft != x.Parent.LFLeft {
			delta := newSizeLeft - x.Parent.SizeLeft
			lfDelta := newLFLeft - x.Parent.LFLeft
			x.Parent.SizeLeft = newSizeLeft
			x.Parent.LFLeft = newLFLeft
			UpdateTreeMetadata(tree, x.Parent, delta, lfDelta)
		}
	}

	RecomputeTreeMetadata(tree, x.Parent)

	if yWasRed {
		ResetSentinel()
		return
	}

	// RB-DELETE-FIXUP
	var w *TreeNode
	for x != tree.Root && x.Color == Black {
		if x == x.Parent.Left {
			w = x.Parent.Right

			if w.Color == Red {
				w.Color = Black
				x.Parent.Color = Red
				LeftRotate(tree, x.Parent)
				w = x.Parent.Right
			}

			if w.Left.Color == Black && w.Right.Color == Black {
				w.Color = Red
				x = x.Parent
			} else {
				if w.Right.Color == Black {
					w.Left.Color = Black
					w.Color = Red
					RightRotate(tree, w)
					w = x.Parent.Right
				}

				w.Color = x.Parent.Color
				x.Parent.Color = Black
				w.Right.Color = Black
				LeftRotate(tree, x.Parent)
				x = tree.Root
			}
		} else {
			w = x.Parent.Left

			if w.Color == Red {
				w.Color = Black
				x.Parent.Color = Red
				RightRotate(tree, x.Parent)
				w = x.Parent.Left
			}

			if w.Left.Color == Black && w.Right.Color == Black {
				w.Color = Red
				x = x.Parent
			} else {
				if w.Left.Color == Black {
					w.Right.Color = Black
					w.Color = Red
					LeftRotate(tree, w)
					w = x.Parent.Left
				}

				w.Color = x.Parent.Color
				x.Parent.Color = Black
				w.Left.Color = Black
				RightRotate(tree, x.Parent)
				x = tree.Root
			}
		}
	}
	x.Color = Black
	ResetSentinel()
}
