package buffer

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// generateLargeText generates a text of specified size in MB
func generateLargeText(sizeMB int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\n\r\t,.;:!?()[]{}\"'+-*/=<>|&^%$#@~`"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Calculate size in bytes
	sizeBytes := sizeMB * 1024 * 1024

	// Use a string builder for better memory efficiency
	var sb strings.Builder
	sb.Grow(sizeBytes)

	// Generate chunks to avoid memory issues
	const chunkSize = 10 * 1024 * 1024 // 10MB chunks
	for i := 0; i < sizeBytes; i += chunkSize {
		size := chunkSize
		if i+size > sizeBytes {
			size = sizeBytes - i
		}

		chunk := make([]byte, size)
		for j := 0; j < size; j++ {
			// Insert a newline every ~80 characters to simulate text
			if j%80 == 0 && j > 0 {
				chunk[j] = '\n'
			} else {
				chunk[j] = chars[r.Intn(len(chars))]
			}
		}
		sb.Write(chunk)
	}

	return sb.String()
}

// generateLargeTextFile creates a temporary file with the specified size in MB
// and returns the file path. The caller is responsible for deleting the file.
func generateLargeTextFile(sizeMB int) (string, error) {
	tempFile, err := os.CreateTemp("", "textbuffer-perf-test-*.txt")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\n\r\t,.;:!?()[]{}\"'+-*/=<>|&^%$#@~`"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Calculate size in bytes
	sizeBytes := sizeMB * 1024 * 1024

	// Generate and write chunks to avoid memory issues
	const chunkSize = 10 * 1024 * 1024 // 10MB chunks
	for i := 0; i < sizeBytes; i += chunkSize {
		size := chunkSize
		if i+size > sizeBytes {
			size = sizeBytes - i
		}

		chunk := make([]byte, size)
		for j := 0; j < size; j++ {
			// Insert a newline every ~80 characters to simulate text
			if j%80 == 0 && j > 0 {
				chunk[j] = '\n'
			} else {
				chunk[j] = chars[r.Intn(len(chars))]
			}
		}

		if _, err := tempFile.Write(chunk); err != nil {
			return tempFile.Name(), err
		}
	}

	return tempFile.Name(), nil
}

// readFileInChunks reads a file in chunks and returns its content
func readFileInChunks(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	fileSize := fileInfo.Size()
	var sb strings.Builder
	sb.Grow(int(fileSize))

	const chunkSize = 10 * 1024 * 1024 // 10MB chunks
	buffer := make([]byte, chunkSize)

	for {
		bytesRead, err := file.Read(buffer)
		if bytesRead > 0 {
			sb.Write(buffer[:bytesRead])
		}

		if err != nil {
			break
		}
	}

	return sb.String(), nil
}

// printMemUsage outputs the current, total and OS memory being used
func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Test1GBFileOperations 测试1GB文件的增删改查性能
func Test1GBFileOperations(t *testing.T) {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 创建临时文件
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("textbuffer-perf-test-%d.txt", rand.Int63()))
	defer os.Remove(tmpFile)

	// 生成1GB的测试文件
	t.Log("Generating 1GB text file...")
	start := time.Now()
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 使用缓冲写入
	buf := make([]byte, 1024*1024) // 1MB buffer
	written := 0
	targetSize := 1024 * 1024 * 1024 // 1GB

	for written < targetSize {
		// 生成随机行
		lineLen := rand.Intn(100) + 1 // 1-100字符
		for i := 0; i < lineLen; i++ {
			buf[i] = byte(rand.Intn(95) + 32) // ASCII 32-126
		}
		buf[lineLen] = '\n'

		// 写入文件
		n, err := file.Write(buf[:lineLen+1])
		if err != nil {
			t.Fatalf("Failed to write to test file: %v", err)
		}
		written += n
	}

	file.Close()
	t.Logf("Generated 1GB text file at: %s", tmpFile)
	t.Logf("Generation took: %v", time.Since(start))

	// 读取文件并创建textbuffer
	t.Log("Reading file and creating text buffer...")
	start = time.Now()
	file, err = os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	// 使用缓冲读取
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024*1024) // 1MB buffer
	chunks := make([]*StringBuffer, 0)
	totalRead := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}
		chunks = append(chunks, NewStringBuffer(string(buffer[:n]), CreateLineStartsFast(string(buffer[:n]), true)))
		totalRead += n
	}

	file.Close()
	t.Logf("Read %d bytes in %v", totalRead, time.Since(start))

	// 创建textbuffer
	t.Log("Creating text buffer...")
	start = time.Now()
	tb := NewPieceTreeBase(chunks, "\n", true)
	t.Logf("Text buffer creation took: %v", time.Since(start))

	// 测试随机访问性能
	t.Log("Testing random access performance...")
	start = time.Now()
	for i := 0; i < 1000; i++ {
		offset := rand.Intn(tb.GetLength())
		_ = tb.GetValueInRange(offset, 1, offset+1, 1, "")
	}
	t.Logf("1000 random accesses took: %v", time.Since(start))

	// 测试行数统计性能
	t.Log("Testing line count performance...")
	start = time.Now()
	lineCount := tb.GetLineCount()
	t.Logf("Line count (%d lines) took: %v", lineCount, time.Since(start))

	// 测试随机行检索性能
	t.Log("Testing random line retrieval performance...")
	start = time.Now()
	for i := 0; i < 100; i++ {
		lineNum := rand.Intn(lineCount) + 1
		_ = tb.GetLineContent(lineNum)
	}
	t.Logf("100 random line retrievals took: %v", time.Since(start))

	// 测试随机编辑性能
	t.Log("Testing random edit performance...")
	start = time.Now()
	for i := 0; i < 100; i++ {
		offset := rand.Intn(tb.GetLength())
		length := rand.Intn(100) + 1
		if offset+length > tb.GetLength() {
			length = tb.GetLength() - offset
		}
		tb.Delete(offset, length)
		tb.Insert(offset, generateRandomString(length), true)
	}
	t.Logf("100 random edits took: %v", time.Since(start))

	// 测试删除操作性能
	t.Log("Testing delete operation performance...")
	start = time.Now()
	// 删除文件中间的一部分
	midPoint := tb.GetLength() / 2
	deleteSize := 1024 * 1024 // 1MB
	tb.Delete(midPoint, deleteSize)
	t.Logf("Delete operation took: %v", time.Since(start))

	// 打印内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Memory usage: Alloc = %v MiB  TotalAlloc = %v MiB  Sys = %v MiB  NumGC = %v",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Test1GBFileDirectOperations 测试1GB文件的直接操作性能
func Test1GBFileDirectOperations(t *testing.T) {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 创建临时文件
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("textbuffer-perf-test-%d.txt", rand.Int63()))
	defer os.Remove(tmpFile)

	// 生成1GB的测试文件
	t.Log("Generating 1GB text file for direct operations...")
	start := time.Now()
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 使用缓冲写入
	buf := make([]byte, 1024*1024) // 1MB buffer
	written := 0
	targetSize := 1024 * 1024 * 1024 // 1GB

	for written < targetSize {
		// 生成随机行
		lineLen := rand.Intn(100) + 1 // 1-100字符
		for i := 0; i < lineLen; i++ {
			buf[i] = byte(rand.Intn(95) + 32) // ASCII 32-126
		}
		buf[lineLen] = '\n'

		// 写入文件
		n, err := file.Write(buf[:lineLen+1])
		if err != nil {
			t.Fatalf("Failed to write to test file: %v", err)
		}
		written += n
	}

	file.Close()
	t.Logf("Generated 1GB text file at: %s", tmpFile)
	t.Logf("Generation took: %v", time.Since(start))

	// 创建textbuffer并直接插入内容
	t.Log("Creating text buffer...")
	start = time.Now()
	tb := NewPieceTreeBase([]*StringBuffer{NewStringBuffer("", []int{0})}, "\n", true)

	// 分块读取并插入
	file, err = os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024*1024) // 1MB buffer
	totalRead := 0
	chunkSize := 100 * 1024 * 1024 // 100MB chunks

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		// 分块插入
		for i := 0; i < n; i += chunkSize {
			end := i + chunkSize
			if end > n {
				end = n
			}
			chunk := string(buffer[i:end])
			tb.Insert(totalRead+i, chunk, true)
		}

		totalRead += n
		if totalRead%chunkSize == 0 {
			t.Logf("Inserted %d/%d MB in %v", totalRead/1024/1024, 1024, time.Since(start))
		}
	}

	file.Close()
	t.Logf("Total insertion of 1GB took: %v", time.Since(start))
	t.Logf("Buffer length: %d bytes", tb.GetLength())

	// 测试随机访问性能
	t.Log("Testing random access performance...")
	start = time.Now()
	for i := 0; i < 1000; i++ {
		offset := rand.Intn(tb.GetLength())
		_ = tb.GetValueInRange(offset, 1, offset+1, 1, "")
	}
	t.Logf("1000 random accesses took: %v", time.Since(start))

	// 测试行数统计性能
	t.Log("Testing line count performance...")
	start = time.Now()
	lineCount := tb.GetLineCount()
	t.Logf("Line count (%d lines) took: %v", lineCount, time.Since(start))

	// 测试随机行检索性能
	t.Log("Testing random line retrieval performance...")
	start = time.Now()
	for i := 0; i < 100; i++ {
		lineNum := rand.Intn(lineCount) + 1
		_ = tb.GetLineContent(lineNum)
	}
	t.Logf("100 random line retrievals took: %v", time.Since(start))

	// 测试随机编辑性能
	t.Log("Testing random edit performance...")
	start = time.Now()
	for i := 0; i < 100; i++ {
		offset := rand.Intn(tb.GetLength())
		length := rand.Intn(100) + 1
		if offset+length > tb.GetLength() {
			length = tb.GetLength() - offset
		}
		tb.Delete(offset, length)
		tb.Insert(offset, generateRandomString(length), true)
	}
	t.Logf("100 random edits took: %v", time.Since(start))

	// 打印内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Memory usage: Alloc = %v MiB  TotalAlloc = %v MiB  Sys = %v MiB  NumGC = %v",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

// TestMediumFileOperations 测试中等大小文件(100MB)的操作性能
func TestMediumFileOperations(t *testing.T) {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 创建临时文件
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("textbuffer-perf-test-%d.txt", rand.Int63()))
	defer os.Remove(tmpFile)

	// 生成100MB的测试文件
	t.Log("Generating 100MB text file for medium file test...")
	start := time.Now()
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 使用缓冲写入
	buf := make([]byte, 1024*1024) // 1MB buffer
	written := 0
	targetSize := 100 * 1024 * 1024 // 100MB

	for written < targetSize {
		// 生成随机行
		lineLen := rand.Intn(100) + 1 // 1-100字符
		for i := 0; i < lineLen; i++ {
			buf[i] = byte(rand.Intn(95) + 32) // ASCII 32-126
		}
		buf[lineLen] = '\n'

		// 写入文件
		n, err := file.Write(buf[:lineLen+1])
		if err != nil {
			t.Fatalf("Failed to write to test file: %v", err)
		}
		written += n
	}

	file.Close()
	t.Logf("Generated 100MB text file at: %s", tmpFile)
	t.Logf("Generation took: %v", time.Since(start))

	// 创建textbuffer并直接插入内容
	t.Log("Creating text buffer...")
	start = time.Now()
	tb := NewPieceTreeBase([]*StringBuffer{NewStringBuffer("", []int{0})}, "\n", true)

	// 分块读取并插入
	file, err = os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024*1024) // 1MB buffer
	totalRead := 0
	chunkSize := 20 * 1024 * 1024 // 20MB chunks

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		// 分块插入
		for i := 0; i < n; i += chunkSize {
			end := i + chunkSize
			if end > n {
				end = n
			}
			chunk := string(buffer[i:end])
			tb.Insert(totalRead+i, chunk, true)
		}

		totalRead += n
		if totalRead%chunkSize == 0 {
			t.Logf("Inserted %d/%d MB", totalRead/1024/1024, 100)
		}
	}

	file.Close()
	t.Logf("Total insertion of 100MB took: %v", time.Since(start))
	t.Logf("Buffer length: %d bytes", tb.GetLength())

	// 测试随机访问性能
	t.Log("Testing random access performance...")
	start = time.Now()
	for i := 0; i < 1000; i++ {
		offset := rand.Intn(tb.GetLength())
		_ = tb.GetValueInRange(offset, 1, offset+1, 1, "")
	}
	t.Logf("1000 random accesses took: %v", time.Since(start))

	// 测试行数统计性能
	t.Log("Testing line count performance...")
	start = time.Now()
	lineCount := tb.GetLineCount()
	t.Logf("Line count (%d lines) took: %v", lineCount, time.Since(start))

	// 测试随机行检索性能
	t.Log("Testing random line retrieval performance...")
	start = time.Now()
	for i := 0; i < 100; i++ {
		lineNum := rand.Intn(lineCount) + 1
		_ = tb.GetLineContent(lineNum)
	}
	t.Logf("100 random line retrievals took: %v", time.Since(start))

	// 测试随机编辑性能
	t.Log("Testing random edit performance...")
	start = time.Now()
	for i := 0; i < 100; i++ {
		offset := rand.Intn(tb.GetLength())
		length := rand.Intn(100) + 1
		if offset+length > tb.GetLength() {
			length = tb.GetLength() - offset
		}
		tb.Delete(offset, length)
		tb.Insert(offset, generateRandomString(length), true)
	}
	t.Logf("100 random edits took: %v", time.Since(start))

	// 打印内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Memory usage: Alloc = %v MiB  TotalAlloc = %v MiB  Sys = %v MiB  NumGC = %v",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}
