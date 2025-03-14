package buffer

// NewPieceTree 从字符串创建一个新的片段树
func NewPieceTree(text string) *PieceTreeBase {
	builder := NewPieceTreeTextBufferBuilder()
	builder.AcceptChunk(text)
	factory := builder.Finish(true)
	return factory.Create(LF)
}
