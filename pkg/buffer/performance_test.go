package buffer

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"testing"
	"time"
)

const (
	// 文件大小定义
	oneMB     = 1024 * 1024
	tenMB     = 10 * oneMB
	hundredMB = 100 * oneMB
	oneGB     = 1024 * oneMB

	// 测试文件路径
	testFilePath = "large_test_file.txt"

	// 操作次数
	numOperations = 1000
)

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// createLargeFile 创建指定大小的测试文件
func createLargeFile(filepath string, size int) error {
	// 每次写入的块大小
	const chunkSize = 10 * oneMB

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	bytesWritten := 0
	for bytesWritten < size {
		writeSize := chunkSize
		if bytesWritten+writeSize > size {
			writeSize = size - bytesWritten
		}

		chunk := generateRandomString(writeSize)
		_, err := file.WriteString(chunk)
		if err != nil {
			return err
		}

		bytesWritten += writeSize
		fmt.Printf("\rCreating test file: %d/%d MB", bytesWritten/oneMB, size/oneMB)
	}
	fmt.Println()

	return nil
}

// reportMemoryUsage 报告当前内存使用情况
func reportMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", m.Alloc/oneMB)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/oneMB)
	fmt.Printf("\tSys = %v MiB", m.Sys/oneMB)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

// PerfTestBuffer 是一个简化的缓冲区，用于性能测试
type PerfTestBuffer struct {
	buffer []byte
}

// NewPerfTestBuffer 创建一个新的性能测试缓冲区
func NewPerfTestBuffer(data []byte) *PerfTestBuffer {
	return &PerfTestBuffer{
		buffer: data,
	}
}

// Read 读取指定范围的内容
func (b *PerfTestBuffer) Read(offset, length int) []byte {
	if offset >= len(b.buffer) {
		return []byte{}
	}
	if offset+length > len(b.buffer) {
		length = len(b.buffer) - offset
	}
	return b.buffer[offset : offset+length]
}

// Insert 在指定位置插入内容
func (b *PerfTestBuffer) Insert(offset int, data []byte) {
	if offset > len(b.buffer) {
		offset = len(b.buffer)
	}

	// 创建新的缓冲区
	newBuffer := make([]byte, len(b.buffer)+len(data))

	// 复制前半部分
	copy(newBuffer[:offset], b.buffer[:offset])

	// 复制插入的数据
	copy(newBuffer[offset:offset+len(data)], data)

	// 复制后半部分
	copy(newBuffer[offset+len(data):], b.buffer[offset:])

	// 更新缓冲区
	b.buffer = newBuffer
}

// Delete 删除指定范围的内容
func (b *PerfTestBuffer) Delete(offset, length int) {
	if offset >= len(b.buffer) {
		return
	}
	if offset+length > len(b.buffer) {
		length = len(b.buffer) - offset
	}

	// 创建新的缓冲区
	newBuffer := make([]byte, len(b.buffer)-length)

	// 复制前半部分
	copy(newBuffer[:offset], b.buffer[:offset])

	// 复制后半部分
	copy(newBuffer[offset:], b.buffer[offset+length:])

	// 更新缓冲区
	b.buffer = newBuffer
}

// TestLargeFileOperations 测试大文件的基本操作
func TestLargeFileOperations(t *testing.T) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 定义测试文件大小
	fileSize := tenMB // 快速测试

	// 创建测试文件
	fmt.Println("Creating test file...")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		err := createLargeFile(testFilePath, fileSize)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// 读取文件内容
	fmt.Println("Reading test file...")
	fileContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// 创建性能测试缓冲区
	perfBuffer := NewPerfTestBuffer(fileContent)
	contentLen := len(fileContent)

	// 测试随机读取
	fmt.Println("\nTesting random reads...")
	startTime := time.Now()
	totalReadBytes := 0
	for i := 0; i < numOperations; i++ {
		offset := rand.Intn(contentLen)
		length := rand.Intn(100) + 1
		if offset+length > contentLen {
			length = contentLen - offset
		}
		if length <= 0 {
			continue
		}

		// 执行读取
		data := perfBuffer.buffer[offset : offset+length]
		totalReadBytes += len(data)
	}
	readDuration := time.Since(startTime)
	fmt.Printf("Random read: %d operations, %d bytes, %v duration\n",
		numOperations, totalReadBytes, readDuration)

	fmt.Println("\nTest completed successfully.")
}

// TestPieceTreePerformance 测试片段树的性能
func TestPieceTreePerformance(t *testing.T) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 定义测试文件大小
	fileSize := tenMB // 快速测试

	// 创建测试文件
	fmt.Println("Creating test file...")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		err := createLargeFile(testFilePath, fileSize)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// 读取文件内容
	fmt.Println("Reading test file...")
	fileContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// 构建片段树缓冲区
	fmt.Println("Building piece tree buffer...")
	startTime := time.Now()
	builder := NewPieceTreeTextBufferBuilder()
	builder.AcceptChunk(string(fileContent))
	factory := builder.Finish(true)
	buffer := factory.Create(LF)
	buildDuration := time.Since(startTime)
	fmt.Printf("Buffer build time: %v\n", buildDuration)
	reportMemoryUsage()

	// 执行随机读取测试
	fmt.Println("\nStarting random read test...")
	startTime = time.Now()
	totalReadLength := 0
	bufferLen := buffer.GetLength()
	for i := 0; i < numOperations; i++ {
		if bufferLen <= 0 {
			break
		}

		startOffset := rand.Intn(bufferLen)
		length := rand.Intn(100) + 1 // 随机读取1到100个字符
		if startOffset+length > bufferLen {
			length = bufferLen - startOffset
		}
		if length <= 0 {
			continue
		}

		// 获取起始位置和结束位置
		startPos := buffer.GetPositionAt(startOffset)
		endPos := buffer.GetPositionAt(startOffset + length)
		if startPos == nil || endPos == nil {
			continue
		}

		// 使用行列信息调用 GetValueInRange
		content := buffer.GetValueInRange(
			startPos.LineNumber,
			startPos.Column,
			endPos.LineNumber,
			endPos.Column,
			buffer.GetEOL())
		totalReadLength += len(content)
	}
	readDuration := time.Since(startTime)
	fmt.Printf("Random read test completed: %d operations, %d bytes total, %v duration\n",
		numOperations, totalReadLength, readDuration)
	reportMemoryUsage()

	fmt.Println("\nTest completed successfully.")
}

// TestComparativePerformance 比较片段树和简单字节数组的性能
func TestComparativePerformance(t *testing.T) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 定义测试文件大小
	fileSize := tenMB
	testFile := "perf_test_file.txt"

	// 创建测试文件
	fmt.Println("Creating test file...")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		err := createLargeFile(testFile, fileSize)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// 读取文件内容
	fmt.Println("Reading test file...")
	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// 创建片段树缓冲区
	fmt.Println("Building piece tree buffer...")
	startTime := time.Now()
	pieceTreeBuilder := NewPieceTreeTextBufferBuilder()
	pieceTreeBuilder.AcceptChunk(string(fileContent))
	factory := pieceTreeBuilder.Finish(true)
	pieceTreeBuffer := factory.Create(LF)
	pieceTreeBuildTime := time.Since(startTime)

	// 创建简单字节数组缓冲区
	fmt.Println("Building byte array buffer...")
	startTime = time.Now()
	byteArrayBuffer := NewPerfTestBuffer(fileContent)
	byteArrayBuildTime := time.Since(startTime)

	fmt.Printf("\nBuild times:\n")
	fmt.Printf("  Piece Tree: %v\n", pieceTreeBuildTime)
	fmt.Printf("  Byte Array: %v\n", byteArrayBuildTime)
	reportMemoryUsage()

	// 执行随机读取测试
	fmt.Println("\n=== Random Read Test ===")
	numReads := 1000

	// 片段树随机读取
	fmt.Println("Piece Tree random read test...")
	startTime = time.Now()
	totalPieceTreeReadLength := 0
	bufferLen := pieceTreeBuffer.GetLength()
	for i := 0; i < numReads; i++ {
		startOffset := rand.Intn(bufferLen)
		length := rand.Intn(100) + 1
		if startOffset+length > bufferLen {
			length = bufferLen - startOffset
		}

		startPos := pieceTreeBuffer.GetPositionAt(startOffset)
		endPos := pieceTreeBuffer.GetPositionAt(startOffset + length)
		if startPos == nil || endPos == nil {
			continue
		}

		content := pieceTreeBuffer.GetValueInRange(
			startPos.LineNumber,
			startPos.Column,
			endPos.LineNumber,
			endPos.Column,
			pieceTreeBuffer.GetEOL())
		totalPieceTreeReadLength += len(content)
	}
	pieceTreeReadTime := time.Since(startTime)

	// 字节数组随机读取
	fmt.Println("Byte Array random read test...")
	startTime = time.Now()
	totalByteArrayReadLength := 0
	contentLen := len(byteArrayBuffer.buffer)
	for i := 0; i < numReads; i++ {
		offset := rand.Intn(contentLen)
		length := rand.Intn(100) + 1
		if offset+length > contentLen {
			length = contentLen - offset
		}

		data := byteArrayBuffer.Read(offset, length)
		totalByteArrayReadLength += len(data)
	}
	byteArrayReadTime := time.Since(startTime)

	fmt.Printf("\nRandom read results (%d operations):\n", numReads)
	fmt.Printf("  Piece Tree: %v, %d bytes\n", pieceTreeReadTime, totalPieceTreeReadLength)
	fmt.Printf("  Byte Array: %v, %d bytes\n", byteArrayReadTime, totalByteArrayReadLength)
	reportMemoryUsage()

	// 执行随机插入测试
	fmt.Println("\n=== Random Insert Test ===")
	numInserts := 100

	// 片段树随机插入
	fmt.Println("Piece Tree random insert test...")
	startTime = time.Now()
	totalPieceTreeInsertLength := 0
	for i := 0; i < numInserts; i++ {
		bufferLen = pieceTreeBuffer.GetLength()
		if bufferLen <= 0 {
			continue
		}
		position := rand.Intn(bufferLen)
		insertText := generateRandomString(rand.Intn(20) + 1) // 插入1到20个字符

		// 直接插入，不需要查询节点
		pieceTreeBuffer.Insert(position, insertText, true)
		totalPieceTreeInsertLength += len(insertText)
	}
	pieceTreeInsertTime := time.Since(startTime)

	// 字节数组随机插入
	fmt.Println("Byte Array random insert test...")
	startTime = time.Now()
	totalByteArrayInsertLength := 0
	for i := 0; i < numInserts; i++ {
		currentLen := len(byteArrayBuffer.buffer)
		offset := rand.Intn(currentLen)
		insertText := generateRandomString(rand.Intn(20) + 1) // 插入1到20个字符

		byteArrayBuffer.Insert(offset, []byte(insertText))
		totalByteArrayInsertLength += len(insertText)
	}
	byteArrayInsertTime := time.Since(startTime)

	fmt.Printf("\nRandom insert results (%d operations):\n", numInserts)
	fmt.Printf("  Piece Tree: %v, %d bytes\n", pieceTreeInsertTime, totalPieceTreeInsertLength)
	fmt.Printf("  Byte Array: %v, %d bytes\n", byteArrayInsertTime, totalByteArrayInsertLength)
	reportMemoryUsage()

	// 执行随机删除测试
	fmt.Println("\n=== Random Delete Test ===")
	numDeletes := 100

	// 字节数组随机删除
	fmt.Println("Byte Array random delete test...")
	startTime = time.Now()
	totalByteArrayDeleteLength := 0
	for i := 0; i < numDeletes; i++ {
		currentLen := len(byteArrayBuffer.buffer)
		if currentLen <= 10 {
			break
		}

		offset := rand.Intn(currentLen - 10)
		length := rand.Intn(10) + 1 // 删除1到10个字符

		byteArrayBuffer.Delete(offset, length)
		totalByteArrayDeleteLength += length
	}
	byteArrayDeleteTime := time.Since(startTime)

	// 片段树随机删除
	fmt.Println("Piece Tree random delete test...")
	startTime = time.Now()
	totalPieceTreeDeleteLength := 0
	for i := 0; i < numDeletes; i++ {
		bufferLen = pieceTreeBuffer.GetLength()
		if bufferLen <= 10 {
			continue
		}

		position := rand.Intn(bufferLen - 10)
		length := rand.Intn(10) + 1 // 删除1到10个字符

		// 直接删除，不需要查询节点
		pieceTreeBuffer.Delete(position, length)
		totalPieceTreeDeleteLength += length
	}
	pieceTreeDeleteTime := time.Since(startTime)

	fmt.Printf("\nRandom delete results (%d operations):\n", numDeletes)
	fmt.Printf("  Piece Tree: %v, %d bytes\n", pieceTreeDeleteTime, totalPieceTreeDeleteLength)
	fmt.Printf("  Byte Array: %v, %d bytes\n", byteArrayDeleteTime, totalByteArrayDeleteLength)
	reportMemoryUsage()

	// 测试完成，删除测试文件
	fmt.Println("\nTest completed successfully.")
	// os.Remove(testFile)
}

// TestSimpleComparative 使用更简单的方式测试片段树性能，只测试读取操作
func TestSimpleComparative(t *testing.T) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 定义测试文件大小 - 使用较小的文件
	fileSize := 1 * 1024 * 1024 // 1MB
	testFile := "perf_test_file_small.txt"

	// 创建测试文件
	fmt.Println("Creating test file...")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		err := createLargeFile(testFile, fileSize)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// 读取文件内容
	fmt.Println("Reading test file...")
	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// 创建片段树缓冲区
	fmt.Println("Building piece tree buffer...")
	startTime := time.Now()
	chunk := NewStringBuffer(string(fileContent), nil)
	chunks := []*StringBuffer{chunk}
	pieceTreeBuffer := NewPieceTreeBase(chunks, "\n", true)
	pieceTreeBuildTime := time.Since(startTime)

	// 创建简单字节数组缓冲区
	fmt.Println("Building byte array buffer...")
	startTime = time.Now()
	byteArrayBuffer := NewPerfTestBuffer(fileContent)
	byteArrayBuildTime := time.Since(startTime)

	fmt.Printf("\nBuild times:\n")
	fmt.Printf("  Piece Tree: %v\n", pieceTreeBuildTime)
	fmt.Printf("  Byte Array: %v\n", byteArrayBuildTime)
	reportMemoryUsage()

	// 执行随机读取测试
	fmt.Println("\n=== Random Read Test ===")
	numReads := 1000

	// 片段树随机读取
	fmt.Println("Piece Tree random read test...")
	startTime = time.Now()
	totalPieceTreeReadLength := 0
	bufferLen := pieceTreeBuffer.GetLength()
	for i := 0; i < numReads; i++ {
		startOffset := rand.Intn(bufferLen)
		length := rand.Intn(100) + 1
		if startOffset+length > bufferLen {
			length = bufferLen - startOffset
		}

		// 使用GetValueInRange方法执行读取
		startPos := pieceTreeBuffer.GetPositionAt(startOffset)
		endPos := pieceTreeBuffer.GetPositionAt(startOffset + length)
		content := pieceTreeBuffer.GetValueInRange(startPos.LineNumber, startPos.Column, endPos.LineNumber, endPos.Column, pieceTreeBuffer.GetEOL())
		totalPieceTreeReadLength += len(content)
	}
	pieceTreeReadTime := time.Since(startTime)

	// 字节数组随机读取
	fmt.Println("Byte Array random read test...")
	startTime = time.Now()
	totalByteArrayReadLength := 0
	contentLen := len(byteArrayBuffer.buffer)
	for i := 0; i < numReads; i++ {
		offset := rand.Intn(contentLen)
		length := rand.Intn(100) + 1
		if offset+length > contentLen {
			length = contentLen - offset
		}

		data := byteArrayBuffer.Read(offset, length)
		totalByteArrayReadLength += len(data)
	}
	byteArrayReadTime := time.Since(startTime)

	fmt.Printf("\nRandom read results (%d operations):\n", numReads)
	fmt.Printf("  Piece Tree: %v, %d bytes\n", pieceTreeReadTime, totalPieceTreeReadLength)
	fmt.Printf("  Byte Array: %v, %d bytes\n", byteArrayReadTime, totalByteArrayReadLength)
	reportMemoryUsage()

	// 执行随机插入测试
	fmt.Println("\n=== Random Insert Test ===")
	numInserts := 100

	// 字节数组随机插入
	fmt.Println("Byte Array random insert test...")
	startTime = time.Now()
	totalByteArrayInsertLength := 0
	byteArrayBufferCopy := NewPerfTestBuffer(fileContent) // 使用副本进行插入测试
	contentLen = len(byteArrayBufferCopy.buffer)
	for i := 0; i < numInserts; i++ {
		offset := rand.Intn(contentLen)
		insertText := generateRandomString(rand.Intn(20) + 1)

		byteArrayBufferCopy.Insert(offset, []byte(insertText))
		totalByteArrayInsertLength += len(insertText)
		contentLen = len(byteArrayBufferCopy.buffer) // 更新长度
	}
	byteArrayInsertTime := time.Since(startTime)

	fmt.Printf("\nRandom insert results (%d operations):\n", numInserts)
	fmt.Printf("  Byte Array: %v, %d bytes\n", byteArrayInsertTime, totalByteArrayInsertLength)
	reportMemoryUsage()

	// 执行随机删除测试
	fmt.Println("\n=== Random Delete Test ===")
	numDeletes := 50

	// 字节数组随机删除
	fmt.Println("Byte Array random delete test...")
	startTime = time.Now()
	totalByteArrayDeleteLength := 0
	contentLen = len(byteArrayBufferCopy.buffer)
	for i := 0; i < numDeletes; i++ {
		if contentLen <= 10 {
			break
		}

		offset := rand.Intn(contentLen - 10)
		length := rand.Intn(10) + 1

		byteArrayBufferCopy.Delete(offset, length)
		totalByteArrayDeleteLength += length
		contentLen = len(byteArrayBufferCopy.buffer) // 更新长度
	}
	byteArrayDeleteTime := time.Since(startTime)

	fmt.Printf("\nRandom delete results (%d operations):\n", numDeletes)
	fmt.Printf("  Byte Array: %v, %d bytes\n", byteArrayDeleteTime, totalByteArrayDeleteLength)
	reportMemoryUsage()

	// 测试完成
	fmt.Println("\nTest completed successfully.")
	// os.Remove(testFile)
}

// TestLargeFilePerformance 测试1GB大文件的性能
func TestLargeFilePerformance(t *testing.T) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 定义测试文件大小
	fileSize := oneGB // 1GB
	testFile := "perf_test_file_1gb.txt"

	// 创建测试文件（分段创建以减少内存压力）
	fmt.Println("Creating 1GB test file...")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		err := createLargeFile(testFile, fileSize)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// 分段处理文件
	fmt.Println("Processing file in chunks...")

	// 片段树构建
	fmt.Println("Building piece tree buffer...")
	startTime := time.Now()
	pieceTreeBuilder := NewPieceTreeTextBufferBuilder()

	// 分段读取文件并构建片段树
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	chunkSize := 100 * oneMB // 每次读取100MB
	buffer := make([]byte, chunkSize)

	bytesRead := 0
	for {
		n, err := file.Read(buffer)
		if err != nil || n == 0 {
			break
		}

		pieceTreeBuilder.AcceptChunk(string(buffer[:n]))
		bytesRead += n
		fmt.Printf("\rReading file for piece tree: %d/%d MB", bytesRead/oneMB, fileSize/oneMB)
	}
	fmt.Println()

	factory := pieceTreeBuilder.Finish(true)
	pieceTreeBuffer := factory.Create(LF)
	pieceTreeBuildTime := time.Since(startTime)

	fmt.Printf("Piece Tree build time: %v\n", pieceTreeBuildTime)
	reportMemoryUsage()

	// 测试读取性能
	fmt.Println("\n=== Random Read Test ===")
	numReads := 1000

	// 片段树随机读取
	fmt.Println("Piece Tree random read test...")
	startTime = time.Now()
	totalPieceTreeReadLength := 0
	bufferLen := pieceTreeBuffer.GetLength()

	for i := 0; i < numReads; i++ {
		startOffset := rand.Intn(bufferLen)
		length := rand.Intn(1000) + 1 // 读取1-1000个字符
		if startOffset+length > bufferLen {
			length = bufferLen - startOffset
		}

		startPos := pieceTreeBuffer.GetPositionAt(startOffset)
		endPos := pieceTreeBuffer.GetPositionAt(startOffset + length)
		if startPos == nil || endPos == nil {
			continue
		}

		content := pieceTreeBuffer.GetValueInRange(
			startPos.LineNumber,
			startPos.Column,
			endPos.LineNumber,
			endPos.Column,
			pieceTreeBuffer.GetEOL())
		totalPieceTreeReadLength += len(content)
	}
	pieceTreeReadTime := time.Since(startTime)
	fmt.Printf("Piece Tree random read: %v, %d bytes\n", pieceTreeReadTime, totalPieceTreeReadLength)
	reportMemoryUsage()

	// 测试插入性能
	fmt.Println("\n=== Random Insert Test ===")
	numInserts := 500

	// 片段树随机插入
	fmt.Println("Piece Tree random insert test...")
	startTime = time.Now()
	totalPieceTreeInsertLength := 0
	insertErrors := 0

	for i := 0; i < numInserts; i++ {
		if i%50 == 0 {
			fmt.Printf("\rInserting: %d/%d", i, numInserts)
		}

		bufferLen = pieceTreeBuffer.GetLength()
		if bufferLen <= 0 {
			continue
		}

		position := rand.Intn(bufferLen)
		insertText := generateRandomString(rand.Intn(50) + 1) // 插入1-50个字符

		// 使用异常处理捕获潜在的空指针异常
		func() {
			defer func() {
				if r := recover(); r != nil {
					insertErrors++
				}
			}()

			pieceTreeBuffer.Insert(position, insertText, true)
			totalPieceTreeInsertLength += len(insertText)
		}()
	}
	fmt.Println()

	pieceTreeInsertTime := time.Since(startTime)
	fmt.Printf("Piece Tree random insert: %v, %d bytes, %d errors\n",
		pieceTreeInsertTime, totalPieceTreeInsertLength, insertErrors)
	reportMemoryUsage()

	// 测试删除性能
	fmt.Println("\n=== Random Delete Test ===")
	numDeletes := 500

	// 片段树随机删除
	fmt.Println("Piece Tree random delete test...")
	startTime = time.Now()
	totalPieceTreeDeleteLength := 0
	deleteErrors := 0

	for i := 0; i < numDeletes; i++ {
		if i%50 == 0 {
			fmt.Printf("\rDeleting: %d/%d", i, numDeletes)
		}

		bufferLen = pieceTreeBuffer.GetLength()
		if bufferLen <= 100 {
			continue
		}

		position := rand.Intn(bufferLen - 100)
		length := rand.Intn(100) + 1 // 删除1-100个字符

		// 使用异常处理捕获潜在的空指针异常
		func() {
			defer func() {
				if r := recover(); r != nil {
					deleteErrors++
				}
			}()

			pieceTreeBuffer.Delete(position, length)
			totalPieceTreeDeleteLength += length
		}()
	}
	fmt.Println()

	pieceTreeDeleteTime := time.Since(startTime)
	fmt.Printf("Piece Tree random delete: %v, %d bytes, %d errors\n",
		pieceTreeDeleteTime, totalPieceTreeDeleteLength, deleteErrors)
	reportMemoryUsage()

	// 添加字节数组的测试代码
	// 为字节数组测试准备较小的测试文件
	smallTestFile := "perf_test_file_sample.txt"
	smallFileSize := 100 * oneMB // 使用100MB作为字节数组测试的样本

	// 创建或使用已有的小文件
	fmt.Println("\n准备字节数组测试样本文件...")
	if _, err := os.Stat(smallTestFile); os.IsNotExist(err) {
		err := createLargeFile(smallTestFile, smallFileSize)
		if err != nil {
			t.Fatalf("Failed to create sample test file: %v", err)
		}
	}

	// 读取样本文件
	fmt.Println("读取样本文件...")
	sampleContent, err := os.ReadFile(smallTestFile)
	if err != nil {
		t.Fatalf("Failed to read sample file: %v", err)
	}

	// 创建字节数组缓冲区
	fmt.Println("创建字节数组缓冲区...")
	startTime = time.Now()
	byteArrayBuffer := NewPerfTestBuffer(sampleContent)
	byteArrayBuildTime := time.Since(startTime)

	fmt.Printf("字节数组构建时间: %v (100MB样本)\n", byteArrayBuildTime)
	reportMemoryUsage()

	// 字节数组随机读取测试
	fmt.Println("\n字节数组随机读取测试...")
	startTime = time.Now()
	totalByteArrayReadLength := 0
	contentLen := len(byteArrayBuffer.buffer)

	for i := 0; i < numReads; i++ {
		offset := rand.Intn(contentLen)
		length := rand.Intn(1000) + 1 // 读取1-1000个字符
		if offset+length > contentLen {
			length = contentLen - offset
		}

		data := byteArrayBuffer.Read(offset, length)
		totalByteArrayReadLength += len(data)
	}

	byteArrayReadTime := time.Since(startTime)
	fmt.Printf("字节数组随机读取: %v, %d bytes (100MB样本)\n", byteArrayReadTime, totalByteArrayReadLength)
	reportMemoryUsage()

	// 字节数组随机插入测试
	fmt.Println("\n字节数组随机插入测试...")
	startTime = time.Now()
	totalByteArrayInsertLength := 0

	for i := 0; i < numInserts; i++ {
		if i%50 == 0 {
			fmt.Printf("\r字节数组插入: %d/%d", i, numInserts)
		}

		currentLen := len(byteArrayBuffer.buffer)
		offset := rand.Intn(currentLen)
		insertText := generateRandomString(rand.Intn(50) + 1) // 插入1-50个字符

		byteArrayBuffer.Insert(offset, []byte(insertText))
		totalByteArrayInsertLength += len(insertText)
	}
	fmt.Println()

	byteArrayInsertTime := time.Since(startTime)
	fmt.Printf("字节数组随机插入: %v, %d bytes (100MB样本)\n", byteArrayInsertTime, totalByteArrayInsertLength)
	reportMemoryUsage()

	// 字节数组随机删除测试
	fmt.Println("\n字节数组随机删除测试...")
	startTime = time.Now()
	totalByteArrayDeleteLength := 0

	for i := 0; i < numDeletes; i++ {
		if i%50 == 0 {
			fmt.Printf("\r字节数组删除: %d/%d", i, numDeletes)
		}

		currentLen := len(byteArrayBuffer.buffer)
		if currentLen <= 100 {
			break
		}

		offset := rand.Intn(currentLen - 100)
		length := rand.Intn(100) + 1 // 删除1-100个字符

		byteArrayBuffer.Delete(offset, length)
		totalByteArrayDeleteLength += length
	}
	fmt.Println()

	byteArrayDeleteTime := time.Since(startTime)
	fmt.Printf("字节数组随机删除: %v, %d bytes (100MB样本)\n", byteArrayDeleteTime, totalByteArrayDeleteLength)
	reportMemoryUsage()

	// 输出性能比较结果
	fmt.Println("\n=== 性能比较结果汇总 ===")
	fmt.Println("构建时间:")
	fmt.Printf("  片段树: %v (1GB文件)\n", pieceTreeBuildTime)
	fmt.Printf("  字节数组: %v (100MB文件)\n", byteArrayBuildTime)

	fmt.Println("\n随机读取 (1000次操作):")
	fmt.Printf("  片段树: %v, %d bytes\n", pieceTreeReadTime, totalPieceTreeReadLength)
	fmt.Printf("  字节数组: %v, %d bytes\n", byteArrayReadTime, totalByteArrayReadLength)

	fmt.Println("\n随机插入 (500次操作):")
	fmt.Printf("  片段树: %v, %d bytes, %d errors\n", pieceTreeInsertTime, totalPieceTreeInsertLength, insertErrors)
	fmt.Printf("  字节数组: %v, %d bytes (100MB样本)\n", byteArrayInsertTime, totalByteArrayInsertLength)

	fmt.Println("\n随机删除 (500次操作):")
	fmt.Printf("  片段树: %v, %d bytes, %d errors\n", pieceTreeDeleteTime, totalPieceTreeDeleteLength, deleteErrors)
	fmt.Printf("  字节数组: %v, %d bytes (100MB样本)\n", byteArrayDeleteTime, totalByteArrayDeleteLength)

	// 测试完成
	fmt.Println("\nTest completed successfully.")
	// 保留文件用于后续分析
	// os.Remove(testFile)
	// os.Remove(smallTestFile)
}

// TestOptimizedLargeFilePerformance 优化版的1GB大文件性能测试
func TestOptimizedLargeFilePerformance(t *testing.T) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 定义测试文件大小 - 使用更小的文件进行全面测试
	fileSize := 100 * oneMB // 使用100MB文件进行全面测试
	testFile := "perf_test_file_100mb.txt"

	// 创建测试文件
	fmt.Println("创建测试文件...")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		err := createLargeFile(testFile, fileSize)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// 读取文件内容
	fmt.Println("读取测试文件...")
	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// 创建片段树缓冲区 - 使用更安全的方式
	fmt.Println("构建片段树缓冲区...")
	startTime := time.Now()
	pieceTreeBuilder := NewPieceTreeTextBufferBuilder()
	pieceTreeBuilder.AcceptChunk(string(fileContent))
	factory := pieceTreeBuilder.Finish(true)
	pieceTreeBuffer := factory.Create(LF)
	pieceTreeBuildTime := time.Since(startTime)

	// 创建字节数组缓冲区
	fmt.Println("构建字节数组缓冲区...")
	startTime = time.Now()
	byteArrayBuffer := NewPerfTestBuffer(fileContent)
	byteArrayBuildTime := time.Since(startTime)

	fmt.Printf("\n构建时间比较:\n")
	fmt.Printf("  片段树: %v\n", pieceTreeBuildTime)
	fmt.Printf("  字节数组: %v\n", byteArrayBuildTime)
	reportMemoryUsage()

	// 执行随机读取测试
	fmt.Println("\n=== 随机读取测试 ===")
	numReads := 10000

	// 片段树随机读取
	fmt.Println("片段树随机读取...")
	startTime = time.Now()
	totalPieceTreeReadLength := 0
	bufferLen := pieceTreeBuffer.GetLength()
	for i := 0; i < numReads; i++ {
		startOffset := rand.Intn(bufferLen)
		length := rand.Intn(100) + 1
		if startOffset+length > bufferLen {
			length = bufferLen - startOffset
		}

		startPos := pieceTreeBuffer.GetPositionAt(startOffset)
		endPos := pieceTreeBuffer.GetPositionAt(startOffset + length)
		if startPos == nil || endPos == nil {
			continue
		}

		content := pieceTreeBuffer.GetValueInRange(
			startPos.LineNumber,
			startPos.Column,
			endPos.LineNumber,
			endPos.Column,
			pieceTreeBuffer.GetEOL())
		totalPieceTreeReadLength += len(content)
	}
	pieceTreeReadTime := time.Since(startTime)

	// 字节数组随机读取
	fmt.Println("字节数组随机读取...")
	startTime = time.Now()
	totalByteArrayReadLength := 0
	contentLen := len(byteArrayBuffer.buffer)
	for i := 0; i < numReads; i++ {
		offset := rand.Intn(contentLen)
		length := rand.Intn(100) + 1
		if offset+length > contentLen {
			length = contentLen - offset
		}

		data := byteArrayBuffer.Read(offset, length)
		totalByteArrayReadLength += len(data)
	}
	byteArrayReadTime := time.Since(startTime)

	fmt.Printf("\n随机读取比较 (%d次操作):\n", numReads)
	fmt.Printf("  片段树: %v, %d bytes\n", pieceTreeReadTime, totalPieceTreeReadLength)
	fmt.Printf("  字节数组: %v, %d bytes\n", byteArrayReadTime, totalByteArrayReadLength)

	// 读取平均时间
	ptReadAvg := float64(pieceTreeReadTime.Nanoseconds()) / float64(numReads)
	baReadAvg := float64(byteArrayReadTime.Nanoseconds()) / float64(numReads)
	fmt.Printf("  单次读取平均时间 - 片段树: %.2f ns, 字节数组: %.2f ns\n", ptReadAvg, baReadAvg)
	reportMemoryUsage()

	// 执行随机插入测试
	fmt.Println("\n=== 随机插入测试 ===")
	numInserts := 1000

	// 创建片段树的副本进行插入测试
	fmt.Println("片段树随机插入...")
	pieceTreeInsertBuffer := factory.Create(LF) // 创建新的实例
	startTime = time.Now()
	totalPieceTreeInsertLength := 0
	insertErrors := 0

	for i := 0; i < numInserts; i++ {
		if i%100 == 0 {
			fmt.Printf("\r片段树插入进度: %d/%d", i, numInserts)
		}

		bufferLen = pieceTreeInsertBuffer.GetLength()
		if bufferLen <= 0 {
			continue
		}

		position := rand.Intn(bufferLen)
		insertText := generateRandomString(rand.Intn(20) + 1) // 插入1-20个字符

		// 使用异常处理捕获潜在的错误
		func() {
			defer func() {
				if r := recover(); r != nil {
					insertErrors++
				}
			}()

			pieceTreeInsertBuffer.Insert(position, insertText, true)
			totalPieceTreeInsertLength += len(insertText)
		}()
	}
	fmt.Println()
	pieceTreeInsertTime := time.Since(startTime)

	// 字节数组随机插入
	fmt.Println("字节数组随机插入...")
	byteArrayInsertBuffer := NewPerfTestBuffer(fileContent) // 创建副本
	startTime = time.Now()
	totalByteArrayInsertLength := 0

	for i := 0; i < numInserts; i++ {
		if i%100 == 0 {
			fmt.Printf("\r字节数组插入进度: %d/%d", i, numInserts)
		}

		currentLen := len(byteArrayInsertBuffer.buffer)
		offset := rand.Intn(currentLen)
		insertText := generateRandomString(rand.Intn(20) + 1)

		byteArrayInsertBuffer.Insert(offset, []byte(insertText))
		totalByteArrayInsertLength += len(insertText)
	}
	fmt.Println()
	byteArrayInsertTime := time.Since(startTime)

	fmt.Printf("\n随机插入比较 (%d次操作):\n", numInserts)
	fmt.Printf("  片段树: %v, %d bytes, %d errors\n", pieceTreeInsertTime, totalPieceTreeInsertLength, insertErrors)
	fmt.Printf("  字节数组: %v, %d bytes\n", byteArrayInsertTime, totalByteArrayInsertLength)

	// 计算平均插入时间
	successfulInserts := numInserts - insertErrors
	if successfulInserts > 0 {
		ptInsertAvg := float64(pieceTreeInsertTime.Nanoseconds()) / float64(successfulInserts)
		baInsertAvg := float64(byteArrayInsertTime.Nanoseconds()) / float64(numInserts)
		fmt.Printf("  单次插入平均时间 - 片段树: %.2f ns, 字节数组: %.2f ns\n", ptInsertAvg, baInsertAvg)
	}
	reportMemoryUsage()

	// 执行随机删除测试
	fmt.Println("\n=== 随机删除测试 ===")
	numDeletes := 1000

	// 片段树随机删除
	fmt.Println("片段树随机删除...")
	startTime = time.Now()
	totalPieceTreeDeleteLength := 0
	deleteErrors := 0

	for i := 0; i < numDeletes; i++ {
		if i%100 == 0 {
			fmt.Printf("\r片段树删除进度: %d/%d", i, numDeletes)
		}

		bufferLen = pieceTreeInsertBuffer.GetLength()
		if bufferLen <= 10 {
			continue
		}

		position := rand.Intn(bufferLen - 10)
		length := rand.Intn(10) + 1

		// 使用异常处理捕获潜在的错误
		func() {
			defer func() {
				if r := recover(); r != nil {
					deleteErrors++
				}
			}()

			pieceTreeInsertBuffer.Delete(position, length)
			totalPieceTreeDeleteLength += length
		}()
	}
	fmt.Println()
	pieceTreeDeleteTime := time.Since(startTime)

	// 字节数组随机删除
	fmt.Println("字节数组随机删除...")
	startTime = time.Now()
	totalByteArrayDeleteLength := 0

	for i := 0; i < numDeletes; i++ {
		if i%100 == 0 {
			fmt.Printf("\r字节数组删除进度: %d/%d", i, numDeletes)
		}

		currentLen := len(byteArrayInsertBuffer.buffer)
		if currentLen <= 10 {
			break
		}

		offset := rand.Intn(currentLen - 10)
		length := rand.Intn(10) + 1

		byteArrayInsertBuffer.Delete(offset, length)
		totalByteArrayDeleteLength += length
	}
	fmt.Println()
	byteArrayDeleteTime := time.Since(startTime)

	fmt.Printf("\n随机删除比较 (%d次操作):\n", numDeletes)
	fmt.Printf("  片段树: %v, %d bytes, %d errors\n", pieceTreeDeleteTime, totalPieceTreeDeleteLength, deleteErrors)
	fmt.Printf("  字节数组: %v, %d bytes\n", byteArrayDeleteTime, totalByteArrayDeleteLength)

	// 计算平均删除时间
	successfulDeletes := numDeletes - deleteErrors
	if successfulDeletes > 0 {
		ptDeleteAvg := float64(pieceTreeDeleteTime.Nanoseconds()) / float64(successfulDeletes)
		baDeleteAvg := float64(byteArrayDeleteTime.Nanoseconds()) / float64(numDeletes)
		fmt.Printf("  单次删除平均时间 - 片段树: %.2f ns, 字节数组: %.2f ns\n", ptDeleteAvg, baDeleteAvg)
	}
	reportMemoryUsage()

	// 测试完成，输出汇总
	fmt.Println("\n=== 性能测试汇总 ===")
	fmt.Println("构建时间:")
	fmt.Printf("  片段树: %v\n", pieceTreeBuildTime)
	fmt.Printf("  字节数组: %v\n", byteArrayBuildTime)

	fmt.Println("\n随机读取:")
	fmt.Printf("  片段树: %v, %.2f ns/op\n", pieceTreeReadTime, ptReadAvg)
	fmt.Printf("  字节数组: %v, %.2f ns/op\n", byteArrayReadTime, baReadAvg)

	if successfulInserts > 0 {
		ptInsertAvg := float64(pieceTreeInsertTime.Nanoseconds()) / float64(successfulInserts)
		baInsertAvg := float64(byteArrayInsertTime.Nanoseconds()) / float64(numInserts)

		fmt.Println("\n随机插入:")
		fmt.Printf("  片段树: %v, %.2f ns/op, 成功率: %d/%d (%.2f%%)\n",
			pieceTreeInsertTime, ptInsertAvg, successfulInserts, numInserts, float64(successfulInserts)/float64(numInserts)*100)
		fmt.Printf("  字节数组: %v, %.2f ns/op\n", byteArrayInsertTime, baInsertAvg)
	}

	if successfulDeletes > 0 {
		ptDeleteAvg := float64(pieceTreeDeleteTime.Nanoseconds()) / float64(successfulDeletes)
		baDeleteAvg := float64(byteArrayDeleteTime.Nanoseconds()) / float64(numDeletes)

		fmt.Println("\n随机删除:")
		fmt.Printf("  片段树: %v, %.2f ns/op, 成功率: %d/%d (%.2f%%)\n",
			pieceTreeDeleteTime, ptDeleteAvg, successfulDeletes, numDeletes, float64(successfulDeletes)/float64(numDeletes)*100)
		fmt.Printf("  字节数组: %v, %.2f ns/op\n", byteArrayDeleteTime, baDeleteAvg)
	}

	fmt.Println("\n测试完成。")
}

// TestFinalBufferPerformance 使用TextBuffer接口测试片段树性能
func TestFinalBufferPerformance(t *testing.T) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 定义测试文件大小和路径
	fileSize := 100 * oneMB
	testFile := "perf_test_file_final.txt"

	// 创建测试文件
	fmt.Println("创建测试文件...")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		err := createLargeFile(testFile, fileSize)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// 读取文件内容
	fmt.Println("读取测试文件...")
	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// 创建片段树文本缓冲区
	fmt.Println("构建片段树缓冲区...")
	startTime := time.Now()
	pieceTreeBuilder := NewPieceTreeTextBufferBuilder()
	pieceTreeBuilder.AcceptChunk(string(fileContent))
	factory := pieceTreeBuilder.Finish(true)
	pieceTreeBuffer := factory.Create(LF)
	pieceTreeBuildTime := time.Since(startTime)

	// 创建字节数组缓冲区
	fmt.Println("构建字节数组缓冲区...")
	startTime = time.Now()
	byteArrayBuffer := NewPerfTestBuffer(fileContent)
	byteArrayBuildTime := time.Since(startTime)

	fmt.Printf("\n构建时间比较:\n")
	fmt.Printf("  片段树: %v\n", pieceTreeBuildTime)
	fmt.Printf("  字节数组: %v\n", byteArrayBuildTime)
	reportMemoryUsage()

	// 执行随机读取测试
	fmt.Println("\n=== 随机读取测试 ===")
	numReads := 10000

	// 片段树随机读取 (使用TextBuffer接口)
	fmt.Println("片段树随机读取测试...")
	startTime = time.Now()
	totalPieceTreeReadLength := 0
	bufferLen := pieceTreeBuffer.GetLength()

	for i := 0; i < numReads; i++ {
		startOffset := rand.Intn(bufferLen)
		length := rand.Intn(100) + 1
		if startOffset+length > bufferLen {
			length = bufferLen - startOffset
		}

		// 使用GetValueInRange函数读取内容
		startPos := pieceTreeBuffer.GetPositionAt(startOffset)
		endPos := pieceTreeBuffer.GetPositionAt(startOffset + length)
		if startPos == nil || endPos == nil {
			continue
		}

		content := pieceTreeBuffer.GetValueInRange(
			startPos.LineNumber,
			startPos.Column,
			endPos.LineNumber,
			endPos.Column,
			pieceTreeBuffer.GetEOL())
		totalPieceTreeReadLength += len(content)
	}
	pieceTreeReadTime := time.Since(startTime)

	// 字节数组随机读取
	fmt.Println("字节数组随机读取测试...")
	startTime = time.Now()
	totalByteArrayReadLength := 0
	contentLen := len(byteArrayBuffer.buffer)

	for i := 0; i < numReads; i++ {
		offset := rand.Intn(contentLen)
		length := rand.Intn(100) + 1
		if offset+length > contentLen {
			length = contentLen - offset
		}

		data := byteArrayBuffer.Read(offset, length)
		totalByteArrayReadLength += len(data)
	}
	byteArrayReadTime := time.Since(startTime)

	fmt.Printf("\n随机读取比较 (%d次操作):\n", numReads)
	fmt.Printf("  片段树: %v, %d bytes\n", pieceTreeReadTime, totalPieceTreeReadLength)
	fmt.Printf("  字节数组: %v, %d bytes\n", byteArrayReadTime, totalByteArrayReadLength)

	// 读取平均时间
	ptReadAvg := float64(pieceTreeReadTime.Nanoseconds()) / float64(numReads)
	baReadAvg := float64(byteArrayReadTime.Nanoseconds()) / float64(numReads)
	fmt.Printf("  单次读取平均时间 - 片段树: %.2f ns, 字节数组: %.2f ns\n", ptReadAvg, baReadAvg)
	reportMemoryUsage()

	fmt.Println("\n测试完成。")
}
