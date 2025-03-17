package buffer

// CacheEntry 缓存条目
type CacheEntry struct {
	// Node 节点
	Node *TreeNode
	// NodeStartOffset 节点起始偏移量
	NodeStartOffset int
	// NodeStartLineNumber 节点起始行号
	NodeStartLineNumber int
}

// PieceTreeSearchCache 片段树搜索缓存
type PieceTreeSearchCache struct {
	// limit 限制
	limit int
	// cache 缓存
	cache []CacheEntry
}

// NewPieceTreeSearchCache 创建一个新的片段树搜索缓存
func NewPieceTreeSearchCache(limit int) *PieceTreeSearchCache {
	return &PieceTreeSearchCache{
		limit: limit,
		cache: make([]CacheEntry, 0),
	}
}

// Get 根据偏移量获取缓存条目
func (c *PieceTreeSearchCache) Get(offset int) *CacheEntry {

	for i := len(c.cache) - 1; i >= 0; i-- {
		if c.cache[i].NodeStartOffset <= offset && c.cache[i].NodeStartOffset+c.cache[i].Node.Piece.Length >= offset {
			return &c.cache[i]
		}
	}
	return nil
}

// Get2 根据行号获取缓存条目
func (c *PieceTreeSearchCache) Get2(lineNumber int) *CacheEntry {
	if len(c.cache) == 0 {
		return nil
	}

	for i := len(c.cache) - 1; i >= 0; i-- {
		if c.cache[i].NodeStartLineNumber > 0 && c.cache[i].NodeStartLineNumber < lineNumber && c.cache[i].NodeStartLineNumber+c.cache[i].Node.Piece.LineFeedCnt >= lineNumber {
			return &c.cache[i]
		}
	}
	return nil
}

// Set 设置缓存条目
func (c *PieceTreeSearchCache) Set(nodePosition CacheEntry) {
	if len(c.cache) >= c.limit {
		c.cache = c.cache[1:]
	}
	c.cache = append(c.cache, nodePosition)
}

// Validate 验证缓存
func (c *PieceTreeSearchCache) Validate(offset int) {
	hasInvalidVal := false
	tmp := make([]*CacheEntry, len(c.cache))
	for i := 0; i < len(c.cache); i++ {
		nodePos := c.cache[i]
		if nodePos.Node.Parent == nil || nodePos.NodeStartOffset >= offset {
			tmp[i] = nil
			hasInvalidVal = true
			continue
		}
		tmp[i] = &nodePos
	}

	if hasInvalidVal {
		newArr := make([]CacheEntry, 0)
		for _, entry := range tmp {
			if entry != nil {
				newArr = append(newArr, *entry)
			}
		}
		c.cache = newArr
	}
}
