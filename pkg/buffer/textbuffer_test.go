package buffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createEmptyTextBuffer() *PieceTreeBase {
	builder := NewPieceTreeTextBufferBuilder()
	factory := builder.Finish(true)
	return factory.Create(LF)
}

func TestTextBufferBasicOperations(t *testing.T) {
	tb := createEmptyTextBuffer()
	assert.Equal(t, 0, tb.GetLength())
	assert.Equal(t, 1, tb.GetLineCount())

	// Insert into empty buffer
	tb.Insert(0, "Hello", true)
	assert.Equal(t, 5, tb.GetLength())
	assert.Equal(t, "Hello", tb.GetLinesRawContent())

	// Append text
	tb.Insert(5, "World", true)
	assert.Equal(t, 10, tb.GetLength())
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())

	// Insert in the middle - this inserts at the end of the first piece
	tb.Insert(5, ", ", true)
	assert.Equal(t, 12, tb.GetLength())
	assert.Equal(t, "Hello, World", tb.GetLinesRawContent())

	// Delete single character
	tb.Delete(5, 1)
	assert.Equal(t, 11, tb.GetLength())
	assert.Equal(t, "Hello World", tb.GetLinesRawContent())

	// Delete multiple characters
	tb.Delete(5, 1)
	assert.Equal(t, 10, tb.GetLength())
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())

	// Insert multiple lines
	tb.Insert(5, "\nWorld\nTest", true)
	assert.Equal(t, 21, tb.GetLength())
	assert.Equal(t, 3, tb.GetLineCount())
	assert.Equal(t, "Hello\nWorld\nTestWorld", tb.GetLinesRawContent())

	// Delete across lines
	tb.Delete(5, 7)
	assert.Equal(t, 14, tb.GetLength())
	assert.Equal(t, 1, tb.GetLineCount())
	assert.Equal(t, "HelloTestWorld", tb.GetLinesRawContent())
}

func TestComprehensiveInsertDelete(t *testing.T) {
	tb := createEmptyTextBuffer()
	assert.Equal(t, 0, tb.GetLength())
	assert.Equal(t, 1, tb.GetLineCount())

	// Insert into empty buffer
	tb.Insert(0, "Hello", true)
	assert.Equal(t, 5, tb.GetLength())
	assert.Equal(t, "Hello", tb.GetLinesRawContent())

	// Append text
	tb.Insert(5, "World", true)
	assert.Equal(t, 10, tb.GetLength())
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())

	// Insert in the middle
	tb.Insert(5, ", ", true)
	assert.Equal(t, 12, tb.GetLength())
	assert.Equal(t, "Hello, World", tb.GetLinesRawContent())

	// Delete single character
	tb.Delete(5, 1)
	assert.Equal(t, 11, tb.GetLength())
	assert.Equal(t, "Hello World", tb.GetLinesRawContent())

	// Delete multiple characters
	tb.Delete(5, 6)
	assert.Equal(t, 5, tb.GetLength())
	assert.Equal(t, "Hello", tb.GetLinesRawContent())

	// Insert multiple lines
	tb.Insert(5, "\nWorld\nTest", true)
	assert.Equal(t, 16, tb.GetLength())
	assert.Equal(t, 3, tb.GetLineCount())
	assert.Equal(t, "Hello\nWorld\nTest", tb.GetLinesRawContent())

	// Delete across lines
	tb.Delete(5, 7)
	assert.Equal(t, 9, tb.GetLength())
	assert.Equal(t, 1, tb.GetLineCount())
	assert.Equal(t, "HelloTest", tb.GetLinesRawContent())

	// CRLF handling
	tb = createEmptyTextBuffer()
	tb.Insert(0, "Line1\r\nLine2\r\nLine3", true)
	assert.Equal(t, 19, tb.GetLength())
	assert.Equal(t, 3, tb.GetLineCount())

	// Delete CRLF
	tb.Delete(5, 2)
	assert.Equal(t, 17, tb.GetLength())
	assert.Equal(t, 2, tb.GetLineCount())
	assert.Equal(t, "Line1Line2\r\nLine3", tb.GetLinesRawContent())

	// Large text operations
	largeText := ""
	for i := 0; i < 1000; i++ {
		largeText += "This is a test line for large text operations.\n"
	}

	tb = createEmptyTextBuffer()
	tb.Insert(0, largeText, true)
	expectedLength := len(largeText)
	assert.Equal(t, expectedLength, tb.GetLength())

	// Delete half of the large text
	halfLength := expectedLength / 2
	tb.Delete(0, halfLength)
	assert.Equal(t, expectedLength-halfLength, tb.GetLength())

	// Multiple small operations
	tb = createEmptyTextBuffer()
	for i := 0; i < 100; i++ {
		tb.Insert(tb.GetLength(), "a", true)
	}
	assert.Equal(t, 100, tb.GetLength())

	for i := 0; i < 50; i++ {
		tb.Delete(0, 1)
	}
	assert.Equal(t, 50, tb.GetLength())

	for i := 0; i < 28; i++ {
		tb.Insert(25, "b", true)
	}
	assert.Equal(t, 78, tb.GetLength())

	// Edge cases
	tb = createEmptyTextBuffer()

	// Insert empty string
	tb.Insert(0, "", true)
	assert.Equal(t, 0, tb.GetLength())

	// Delete zero characters
	tb.Delete(0, 0)
	assert.Equal(t, 0, tb.GetLength())

	// Insert at out of bounds
	tb.Insert(0, "Test", true)
	assert.Equal(t, 4, tb.GetLength())

	// Delete out of bounds should be handled gracefully
	tb.Delete(10, 5)
	assert.Equal(t, 4, tb.GetLength())

	// Delete more than available
	tb.Delete(0, 10)
	assert.Equal(t, 0, tb.GetLength())
}

func TestSimpleDelete(t *testing.T) {
	tb := createEmptyTextBuffer()

	// Insert "Hello, World"
	tb.Insert(0, "Hello, World", true)
	assert.Equal(t, 12, tb.GetLength())
	assert.Equal(t, "Hello, World", tb.GetLinesRawContent())

	t.Log("Buffer structure after insert:", len(tb.buffers), "buffers")
	t.Logf("Buffer 0: '%s', lineStarts: %v", tb.buffers[0].Buffer, tb.buffers[0].LineStarts)

	// Delete the comma
	tb.Delete(5, 1)
	assert.Equal(t, 11, tb.GetLength())
	assert.Equal(t, "Hello World", tb.GetLinesRawContent())
	t.Log("Buffer structure after delete:", len(tb.buffers), "buffers")
	t.Logf("Buffer 0: '%s', lineStarts: %v", tb.buffers[0].Buffer, tb.buffers[0].LineStarts)

	// Delete the space
	tb.Delete(5, 1)
	assert.Equal(t, 10, tb.GetLength())
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())
	t.Log("Buffer structure after second delete:", len(tb.buffers), "buffers")
	t.Logf("Buffer 0: '%s', lineStarts: %v", tb.buffers[0].Buffer, tb.buffers[0].LineStarts)

	// Test with a different insertion order
	tb = createEmptyTextBuffer()

	// Insert "Hello" first
	tb.Insert(0, "Hello", true)
	// Insert "World" next
	tb.Insert(5, "World", true)
	// Insert ", " in the middle
	tb.Insert(5, ", ", true)

	t.Log("Buffer structure before delete in second test:", len(tb.buffers), "buffers")
	t.Logf("Buffer 0: '%s', lineStarts: %v", tb.buffers[0].Buffer, tb.buffers[0].LineStarts)

	// Log the tree structure
	t.Log("Tree structure before delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		t.Log("  SENTINEL node")
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s'\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node))
		}
		return true
	})

	// Log node at offset 5
	pos := tb.NodeAt(5)
	if pos.Node != nil {
		t.Logf("Node at offset 5: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', remainder=%d, nodeStartOffset=%d",
			pos.Node.Piece.BufferIndex,
			pos.Node.Piece.Start.Line, pos.Node.Piece.Start.Column,
			pos.Node.Piece.End.Line, pos.Node.Piece.End.Column,
			pos.Node.Piece.Length,
			tb.GetNodeContent(pos.Node),
			pos.Remainder,
			pos.NodeStartOffset)
	}

	// Log node at offset 6
	pos = tb.NodeAt(6)
	if pos.Node != nil {
		t.Logf("Node at offset 6: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', remainder=%d, nodeStartOffset=%d",
			pos.Node.Piece.BufferIndex,
			pos.Node.Piece.Start.Line, pos.Node.Piece.Start.Column,
			pos.Node.Piece.End.Line, pos.Node.Piece.End.Column,
			pos.Node.Piece.Length,
			tb.GetNodeContent(pos.Node),
			pos.Remainder,
			pos.NodeStartOffset)
	}

	// Delete the comma
	tb.Delete(5, 1)
	assert.Equal(t, 11, tb.GetLength())
	assert.Equal(t, "Hello World", tb.GetLinesRawContent())
	t.Log("Buffer structure after delete in second test:", len(tb.buffers), "buffers")
	t.Logf("Buffer 0: '%s', lineStarts: %v", tb.buffers[0].Buffer, tb.buffers[0].LineStarts)

	// Log the tree structure after delete
	t.Log("Tree structure after delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		t.Log("  SENTINEL node")
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s'\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node))
		}
		return true
	})

	assert.Equal(t, "Hello World", tb.GetLinesRawContent())

	// Delete the space
	tb.Delete(5, 1)
	assert.Equal(t, 10, tb.GetLength())
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())
	t.Log("Buffer structure after second delete in second test:", len(tb.buffers), "buffers")
	t.Logf("Buffer 0: '%s', lineStarts: %v", tb.buffers[0].Buffer, tb.buffers[0].LineStarts)

	// Log the tree structure after second delete
	t.Log("Tree structure after second delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		t.Log("  SENTINEL node")
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s'\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node))
		}
		return true
	})

	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())
}

func TestDeleteIssue(t *testing.T) {
	// Test case 1: Insert in order "Hello", ", ", "World"
	tb1 := createEmptyTextBuffer()
	tb1.Insert(0, "Hello", true)
	tb1.Insert(5, ", ", true)
	tb1.Insert(7, "World", true)

	assert.Equal(t, 12, tb1.GetLength())
	assert.Equal(t, "Hello, World", tb1.GetLinesRawContent())

	// Delete the comma and space
	tb1.Delete(5, 2)
	assert.Equal(t, 10, tb1.GetLength())
	assert.Equal(t, "HelloWorld", tb1.GetLinesRawContent())

	// Test case 2: Insert in order "Hello", "World", ", " (in the middle)
	tb2 := createEmptyTextBuffer()
	tb2.Insert(0, "Hello", true)
	tb2.Insert(5, "World", true)
	tb2.Insert(5, ", ", true)

	assert.Equal(t, 12, tb2.GetLength())
	assert.Equal(t, "Hello, World", tb2.GetLinesRawContent())

	// Log the tree structure before delete
	t.Log("Tree structure before delete (tb2):")
	tb2.Iterate(tb2.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s'\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb2.GetNodeContent(node))
		}
		return true
	})

	// Log node at offset 5
	pos := tb2.NodeAt(5)
	if pos.Node != nil {
		t.Logf("Node at offset 5: bufferIndex=%d, content='%s', remainder=%d, nodeStartOffset=%d",
			pos.Node.Piece.BufferIndex,
			tb2.GetNodeContent(pos.Node),
			pos.Remainder,
			pos.NodeStartOffset)
	}

	// Log node at offset 7
	pos = tb2.NodeAt(7)
	if pos.Node != nil {
		t.Logf("Node at offset 7: bufferIndex=%d, content='%s', remainder=%d, nodeStartOffset=%d",
			pos.Node.Piece.BufferIndex,
			tb2.GetNodeContent(pos.Node),
			pos.Remainder,
			pos.NodeStartOffset)
	}

	// Log the raw content before delete
	t.Logf("Raw content before delete: '%s'", tb2.GetLinesRawContent())

	// Delete the comma and space
	tb2.Delete(5, 2)

	// Log the raw content after delete
	t.Logf("Raw content after delete: '%s'", tb2.GetLinesRawContent())

	// Log the tree structure after delete
	t.Log("Tree structure after delete (tb2):")
	tb2.Iterate(tb2.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s'\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb2.GetNodeContent(node))
		}
		return true
	})

	assert.Equal(t, 10, tb2.GetLength())
	assert.Equal(t, "HelloWorld", tb2.GetLinesRawContent())
}

func TestDeleteSpecificIssue(t *testing.T) {
	// Create a buffer with "Hello, World"
	tb := createEmptyTextBuffer()
	tb.Insert(0, "Hello", true)
	tb.Insert(5, "World", true)
	tb.Insert(5, ", ", true)

	assert.Equal(t, 12, tb.GetLength())
	assert.Equal(t, "Hello, World", tb.GetLinesRawContent())

	// Log the tree structure
	t.Log("Tree structure before delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', SizeLeft=%d, LFLeft=%d\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node),
				node.SizeLeft,
				node.LFLeft)
		}
		return true
	})

	// Log the node at offset 5
	pos := tb.NodeAt(5)
	if pos.Node != nil {
		t.Logf("Node at offset 5: bufferIndex=%d, content='%s', remainder=%d, nodeStartOffset=%d",
			pos.Node.Piece.BufferIndex,
			tb.GetNodeContent(pos.Node),
			pos.Remainder,
			pos.NodeStartOffset)
	}

	// Log the node at offset 6
	pos = tb.NodeAt(6)
	if pos.Node != nil {
		t.Logf("Node at offset 6: bufferIndex=%d, content='%s', remainder=%d, nodeStartOffset=%d",
			pos.Node.Piece.BufferIndex,
			tb.GetNodeContent(pos.Node),
			pos.Remainder,
			pos.NodeStartOffset)
	}

	// Log the node at offset 7
	pos = tb.NodeAt(7)
	if pos.Node != nil {
		t.Logf("Node at offset 7: bufferIndex=%d, content='%s', remainder=%d, nodeStartOffset=%d",
			pos.Node.Piece.BufferIndex,
			tb.GetNodeContent(pos.Node),
			pos.Remainder,
			pos.NodeStartOffset)
	}

	// Try to delete just the comma
	tb.Delete(5, 1)

	// Log the tree structure after delete
	t.Log("Tree structure after deleting comma:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', SizeLeft=%d, LFLeft=%d\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node),
				node.SizeLeft,
				node.LFLeft)
		}
		return true
	})

	// Log the content after deleting comma
	t.Logf("Content after deleting comma: '%s'", tb.GetLinesRawContent())

	// Try to delete the space
	tb.Delete(5, 1)

	// Log the tree structure after delete
	t.Log("Tree structure after deleting space:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', SizeLeft=%d, LFLeft=%d\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node),
				node.SizeLeft,
				node.LFLeft)
		}
		return true
	})

	// Log the content after deleting space
	t.Logf("Content after deleting space: '%s'", tb.GetLinesRawContent())

	// Final assertion
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())
}

func TestNodeBoundaryDelete(t *testing.T) {
	// Create a buffer with "Hello, World"
	tb := createEmptyTextBuffer()
	tb.Insert(0, "Hello", true)
	tb.Insert(5, ", ", true)
	tb.Insert(7, "World", true)

	assert.Equal(t, 12, tb.GetLength())
	assert.Equal(t, "Hello, World", tb.GetLinesRawContent())

	// Log the tree structure
	t.Log("Tree structure before delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', SizeLeft=%d, LFLeft=%d\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node),
				node.SizeLeft,
				node.LFLeft)
		}
		return true
	})

	// Delete the comma and space
	tb.Delete(5, 2)

	// Log the tree structure after delete
	t.Log("Tree structure after delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', SizeLeft=%d, LFLeft=%d\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node),
				node.SizeLeft,
				node.LFLeft)
		}
		return true
	})

	// Log the content after delete
	t.Logf("Content after delete: '%s'", tb.GetLinesRawContent())

	// Final assertion
	assert.Equal(t, 10, tb.GetLength())
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())
}

func TestNodeBoundaryDeleteWithDifferentInsertOrder(t *testing.T) {
	// Create a buffer with "Hello, World" but with a different insertion order
	tb := createEmptyTextBuffer()
	tb.Insert(0, "Hello", true)
	tb.Insert(5, "World", true)
	tb.Insert(5, ", ", true)

	assert.Equal(t, 12, tb.GetLength())
	assert.Equal(t, "Hello, World", tb.GetLinesRawContent())

	// Log the tree structure
	t.Log("Tree structure before delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', SizeLeft=%d, LFLeft=%d\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node),
				node.SizeLeft,
				node.LFLeft)
		}
		return true
	})

	// Log nodes at specific offsets
	for i := 0; i <= 12; i++ {
		pos := tb.NodeAt(i)
		if pos.Node != nil {
			t.Logf("Node at offset %d: bufferIndex=%d, content='%s', remainder=%d, nodeStartOffset=%d",
				i,
				pos.Node.Piece.BufferIndex,
				tb.GetNodeContent(pos.Node),
				pos.Remainder,
				pos.NodeStartOffset)
		}
	}

	// Delete the comma and space
	tb.Delete(5, 2)

	// Log the tree structure after delete
	t.Log("Tree structure after delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s', SizeLeft=%d, LFLeft=%d\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node),
				node.SizeLeft,
				node.LFLeft)
		}
		return true
	})

	// Log the content after delete
	t.Logf("Content after delete: '%s'", tb.GetLinesRawContent())

	// Final assertion
	assert.Equal(t, 10, tb.GetLength())
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())
}

func TestDeleteCommaSpace(t *testing.T) {
	// Create a buffer with "Hello, World" using the problematic insertion order
	tb := createEmptyTextBuffer()
	tb.Insert(0, "Hello", true)
	tb.Insert(5, "World", true)
	tb.Insert(5, ", ", true)

	assert.Equal(t, 12, tb.GetLength())
	assert.Equal(t, "Hello, World", tb.GetLinesRawContent())

	// Log the tree structure
	t.Log("Tree structure before delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s'\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node))
		}
		return true
	})

	// Log the node at offset 5
	pos := tb.NodeAt(5)
	if pos.Node != nil {
		t.Logf("Node at offset 5: bufferIndex=%d, content='%s', remainder=%d, nodeStartOffset=%d",
			pos.Node.Piece.BufferIndex,
			tb.GetNodeContent(pos.Node),
			pos.Remainder,
			pos.NodeStartOffset)
	}

	// Manually find and delete the comma and space node
	var commaSpaceNode *TreeNode
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL && tb.GetNodeContent(node) == ", " {
			commaSpaceNode = node
			return false
		}
		return true
	})

	if commaSpaceNode != nil {
		t.Logf("Found comma space node: content='%s'", tb.GetNodeContent(commaSpaceNode))
		RbDelete(tb, commaSpaceNode)
	} else {
		t.Error("Could not find comma space node")
	}

	// Log the tree structure after delete
	t.Log("Tree structure after delete:")
	tb.Iterate(tb.Root, func(node *TreeNode) bool {
		if node != SENTINEL {
			t.Logf("  Node: bufferIndex=%d, start={%d %d}, end={%d %d}, length=%d, content='%s'\n",
				node.Piece.BufferIndex,
				node.Piece.Start.Line, node.Piece.Start.Column,
				node.Piece.End.Line, node.Piece.End.Column,
				node.Piece.Length,
				tb.GetNodeContent(node))
		}
		return true
	})

	// Log the content after delete
	t.Logf("Content after delete: '%s'", tb.GetLinesRawContent())

	// Final assertion
	assert.Equal(t, "HelloWorld", tb.GetLinesRawContent())
}
