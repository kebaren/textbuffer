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

	// Delete multiple characters
	tb.Delete(5, 1)
	assert.Equal(t, 10, tb.GetLength())
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
	assert.Equal(t, "Hellod", tb.GetLinesRawContent())

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
