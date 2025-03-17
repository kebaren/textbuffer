package buffer

import (
	"fmt"
	"math/rand"
	"os"
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

func Test1GBFileOperations(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("skipping 1GB file test in short mode")
	}

	// Use file-based approach to avoid memory issues
	const sizeMB = 1024 // 1GB
	fmt.Printf("Generating %dMB text file...\n", sizeMB)

	filePath, err := generateLargeTextFile(sizeMB)
	if err != nil {
		t.Fatalf("Failed to generate large text file: %v", err)
	}
	defer os.Remove(filePath) // Clean up the file when done

	fmt.Printf("Generated %dMB text file at: %s\n", sizeMB, filePath)
	printMemUsage()

	// Read the file in chunks
	fmt.Println("Reading file in chunks...")
	start := time.Now()
	content, err := readFileInChunks(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	readDuration := time.Since(start)
	fmt.Printf("Read %d bytes in %v\n", len(content), readDuration)
	printMemUsage()

	// Create and populate the text buffer
	fmt.Println("Creating text buffer and inserting content...")
	tb := createEmptyTextBuffer()

	// Insert content in chunks to avoid memory issues
	const insertChunkSize = 10 * 1024 * 1024 // 10MB chunks
	insertStart := time.Now()

	for i := 0; i < len(content); i += insertChunkSize {
		end := i + insertChunkSize
		if end > len(content) {
			end = len(content)
		}

		chunkStart := time.Now()
		tb.Insert(i, content[i:end], true)
		chunkDuration := time.Since(chunkStart)

		if (i/insertChunkSize)%10 == 0 { // Report every 10 chunks (100MB)
			fmt.Printf("Inserted %d/%d MB in %v\n", end/(1024*1024), sizeMB, chunkDuration)
			printMemUsage()
		}
	}

	insertDuration := time.Since(insertStart)
	fmt.Printf("Total insertion of %dMB took: %v\n", sizeMB, insertDuration)
	printMemUsage()

	// Verify buffer size
	bufferLength := tb.GetLength()
	fmt.Printf("Buffer length: %d bytes\n", bufferLength)
	if bufferLength != len(content) {
		t.Errorf("Buffer length mismatch: expected %d, got %d", len(content), bufferLength)
	}

	// Measure random access time
	fmt.Println("Performing random access operations...")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	accessStart := time.Now()

	for j := 0; j < 1000; j++ {
		offset := r.Intn(tb.GetLength())
		tb.NodeAt(offset)
	}

	accessDuration := time.Since(accessStart)
	fmt.Printf("1000 random accesses took: %v\n", accessDuration)
	printMemUsage()

	// Measure line count
	lineCountStart := time.Now()
	lineCount := tb.GetLineCount()
	lineCountDuration := time.Since(lineCountStart)
	fmt.Printf("Line count (%d lines) took: %v\n", lineCount, lineCountDuration)

	// Measure random line retrievals
	fmt.Println("Performing random line retrievals...")
	lineStart := time.Now()

	for j := 0; j < 100; j++ {
		lineNum := r.Intn(lineCount) + 1 // Lines are 1-indexed
		tb.GetLineContent(lineNum)
	}

	lineDuration := time.Since(lineStart)
	fmt.Printf("100 random line retrievals took: %v\n", lineDuration)
	printMemUsage()

	// Measure random small edits
	fmt.Println("Performing random edits...")
	editStart := time.Now()

	for j := 0; j < 100; j++ {
		offset := r.Intn(tb.GetLength())
		length := r.Intn(10) + 1 // Delete 1-10 characters
		if offset+length > tb.GetLength() {
			length = tb.GetLength() - offset
		}
		if length > 0 {
			tb.Delete(offset, length)
			tb.Insert(offset, "replacement", true)
		}
	}

	editDuration := time.Since(editStart)
	fmt.Printf("100 random edits took: %v\n", editDuration)
	printMemUsage()

	// Summary
	fmt.Println("\nPerformance Summary:")
	fmt.Printf("File Size: %d MB\n", sizeMB)
	fmt.Printf("Read Time: %v\n", readDuration)
	fmt.Printf("Insert Time: %v\n", insertDuration)
	fmt.Printf("Random Access Time (1000 ops): %v\n", accessDuration)
	fmt.Printf("Line Count Time: %v\n", lineCountDuration)
	fmt.Printf("Line Retrieval Time (100 ops): %v\n", lineDuration)
	fmt.Printf("Edit Time (100 ops): %v\n", editDuration)
}

func BenchmarkTextBufferLargeFile(b *testing.B) {
	// Skip in short mode
	if testing.Short() {
		b.Skip("skipping large file test in short mode")
	}

	// Generate a large text (100MB for testing, can be increased)
	// Note: For actual 1GB test, change this to 1024
	const sizeMB = 100
	fmt.Printf("Generating %dMB of text...\n", sizeMB)
	largeText := generateLargeText(sizeMB)
	fmt.Printf("Generated %d bytes of text\n", len(largeText))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tb := createEmptyTextBuffer()
		b.StartTimer()

		// Measure insertion time
		insertStart := time.Now()
		tb.Insert(0, largeText, true)
		insertDuration := time.Since(insertStart)
		fmt.Printf("Insertion of %dMB took: %v\n", sizeMB, insertDuration)

		// Measure random access time
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		accessStart := time.Now()
		for j := 0; j < 1000; j++ {
			offset := r.Intn(tb.GetLength())
			tb.NodeAt(offset)
		}
		accessDuration := time.Since(accessStart)
		fmt.Printf("1000 random accesses took: %v\n", accessDuration)

		// Measure random small edits
		editStart := time.Now()
		for j := 0; j < 100; j++ {
			offset := r.Intn(tb.GetLength())
			length := r.Intn(10) + 1 // Delete 1-10 characters
			if offset+length > tb.GetLength() {
				length = tb.GetLength() - offset
			}
			if length > 0 {
				tb.Delete(offset, length)
				tb.Insert(offset, "replacement", true)
			}
		}
		editDuration := time.Since(editStart)
		fmt.Printf("100 random edits took: %v\n", editDuration)

		// Measure line count and content retrieval
		lineStart := time.Now()
		lineCount := tb.GetLineCount()
		for j := 0; j < 100; j++ {
			lineNum := r.Intn(lineCount) + 1 // Lines are 1-indexed
			tb.GetLineContent(lineNum)
		}
		lineDuration := time.Since(lineStart)
		fmt.Printf("Line count and 100 random line retrievals took: %v\n", lineDuration)
	}
}

func BenchmarkTextBufferIncrementalEdits(b *testing.B) {
	// Skip in short mode
	if testing.Short() {
		b.Skip("skipping incremental edits test in short mode")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tb := createEmptyTextBuffer()

		// Measure incremental growth
		const totalOps = 100000
		const reportInterval = 10000

		fmt.Println("Starting incremental edit test...")
		start := time.Now()

		for j := 0; j < totalOps; j++ {
			// Insert a line of text
			tb.Insert(tb.GetLength(), fmt.Sprintf("Line %d of text for testing incremental edits\n", j), true)

			// Every 100 operations, do some random edits
			if j%100 == 0 && tb.GetLength() > 0 {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				offset := r.Intn(tb.GetLength())
				length := r.Intn(10) + 1
				if offset+length > tb.GetLength() {
					length = tb.GetLength() - offset
				}
				if length > 0 {
					tb.Delete(offset, length)
				}
			}

			// Report progress
			if (j+1)%reportInterval == 0 {
				elapsed := time.Since(start)
				fmt.Printf("%d operations completed in %v (%.2f ops/sec), buffer size: %d bytes\n",
					j+1, elapsed, float64(j+1)/elapsed.Seconds(), tb.GetLength())
			}
		}

		totalTime := time.Since(start)
		fmt.Printf("Completed %d incremental operations in %v (%.2f ops/sec)\n",
			totalOps, totalTime, float64(totalOps)/totalTime.Seconds())
		fmt.Printf("Final buffer size: %d bytes, line count: %d\n", tb.GetLength(), tb.GetLineCount())
	}
}

func BenchmarkTextBufferNodeBoundaryOperations(b *testing.B) {
	// This benchmark specifically tests operations at node boundaries
	// which was the source of the bug we fixed

	for i := 0; i < b.N; i++ {
		tb := createEmptyTextBuffer()

		// Insert initial content with different patterns
		tb.Insert(0, "Hello", true)
		tb.Insert(5, "World", true)
		tb.Insert(5, ", ", true)

		// Benchmark node boundary operations
		start := time.Now()

		// Perform many operations at the boundary between nodes
		for j := 0; j < 10000; j++ {
			// Delete at the boundary
			tb.Delete(5, 2)
			// Insert at the boundary
			tb.Insert(5, ", ", true)
		}

		duration := time.Since(start)
		fmt.Printf("10000 node boundary operations took: %v (%.2f ops/sec)\n",
			duration, float64(20000)/duration.Seconds())
	}
}

// Test1GBFileDirectOperations tests operations on a 1GB file without loading the entire content into memory
func Test1GBFileDirectOperations(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("skipping 1GB direct file test in short mode")
	}

	// Use file-based approach to avoid memory issues
	const sizeMB = 1024 // 1GB
	fmt.Printf("Generating %dMB text file for direct operations...\n", sizeMB)

	filePath, err := generateLargeTextFile(sizeMB)
	if err != nil {
		t.Fatalf("Failed to generate large text file: %v", err)
	}
	defer os.Remove(filePath) // Clean up the file when done

	fmt.Printf("Generated %dMB text file at: %s\n", sizeMB, filePath)
	printMemUsage()

	// Create text buffer
	fmt.Println("Creating text buffer...")
	tb := createEmptyTextBuffer()

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Insert content in chunks directly from file to avoid memory issues
	const insertChunkSize = 10 * 1024 * 1024 // 10MB chunks
	buffer := make([]byte, insertChunkSize)

	insertStart := time.Now()
	totalBytesRead := 0

	for {
		bytesRead, err := file.Read(buffer)
		if bytesRead > 0 {
			chunkStart := time.Now()
			tb.Insert(totalBytesRead, string(buffer[:bytesRead]), true)
			chunkDuration := time.Since(chunkStart)

			totalBytesRead += bytesRead

			if (totalBytesRead/(1024*1024))%100 == 0 { // Report every 100MB
				fmt.Printf("Inserted %d/%d MB in %v\n", totalBytesRead/(1024*1024), sizeMB, chunkDuration)
				printMemUsage()

				// Force garbage collection to free memory
				runtime.GC()
			}
		}

		if err != nil {
			break
		}
	}

	insertDuration := time.Since(insertStart)
	fmt.Printf("Total insertion of %dMB took: %v\n", sizeMB, insertDuration)
	printMemUsage()

	// Verify buffer size
	bufferLength := tb.GetLength()
	fmt.Printf("Buffer length: %d bytes\n", bufferLength)
	expectedSize := int64(sizeMB) * 1024 * 1024
	if int64(bufferLength) != expectedSize {
		t.Errorf("Buffer length mismatch: expected %d, got %d", expectedSize, bufferLength)
	}

	// Measure random access time
	fmt.Println("Performing random access operations...")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	accessStart := time.Now()

	for j := 0; j < 1000; j++ {
		offset := r.Intn(tb.GetLength())
		tb.NodeAt(offset)

		// Force GC every 100 operations to prevent memory issues
		if j%100 == 0 {
			runtime.GC()
		}
	}

	accessDuration := time.Since(accessStart)
	fmt.Printf("1000 random accesses took: %v\n", accessDuration)
	printMemUsage()

	// Measure line count
	lineCountStart := time.Now()
	lineCount := tb.GetLineCount()
	lineCountDuration := time.Since(lineCountStart)
	fmt.Printf("Line count (%d lines) took: %v\n", lineCount, lineCountDuration)

	// Measure random line retrievals
	fmt.Println("Performing random line retrievals...")
	lineStart := time.Now()

	for j := 0; j < 100; j++ {
		lineNum := r.Intn(lineCount) + 1 // Lines are 1-indexed
		tb.GetLineContent(lineNum)

		// Force GC every 10 operations to prevent memory issues
		if j%10 == 0 {
			runtime.GC()
		}
	}

	lineDuration := time.Since(lineStart)
	fmt.Printf("100 random line retrievals took: %v\n", lineDuration)
	printMemUsage()

	// Measure random small edits
	fmt.Println("Performing random edits...")
	editStart := time.Now()

	for j := 0; j < 100; j++ {
		offset := r.Intn(tb.GetLength())
		length := r.Intn(10) + 1 // Delete 1-10 characters
		if offset+length > tb.GetLength() {
			length = tb.GetLength() - offset
		}
		if length > 0 {
			tb.Delete(offset, length)
			tb.Insert(offset, "replacement", true)
		}

		// Force GC every 10 operations to prevent memory issues
		if j%10 == 0 {
			runtime.GC()
		}
	}

	editDuration := time.Since(editStart)
	fmt.Printf("100 random edits took: %v\n", editDuration)
	printMemUsage()

	// Summary
	fmt.Println("\nPerformance Summary (Direct File Operations):")
	fmt.Printf("File Size: %d MB\n", sizeMB)
	fmt.Printf("Insert Time: %v\n", insertDuration)
	fmt.Printf("Random Access Time (1000 ops): %v\n", accessDuration)
	fmt.Printf("Line Count Time: %v\n", lineCountDuration)
	fmt.Printf("Line Retrieval Time (100 ops): %v\n", lineDuration)
	fmt.Printf("Edit Time (100 ops): %v\n", editDuration)
}

// TestMediumFileOperations tests operations on a 100MB file for quicker testing
func TestMediumFileOperations(t *testing.T) {
	// This test is designed to run more quickly than the 1GB test
	// but still test performance with a reasonably large file

	// Use file-based approach to avoid memory issues
	const sizeMB = 100 // 100MB
	fmt.Printf("Generating %dMB text file for medium file test...\n", sizeMB)

	filePath, err := generateLargeTextFile(sizeMB)
	if err != nil {
		t.Fatalf("Failed to generate medium text file: %v", err)
	}
	defer os.Remove(filePath) // Clean up the file when done

	fmt.Printf("Generated %dMB text file at: %s\n", sizeMB, filePath)
	printMemUsage()

	// Create text buffer
	fmt.Println("Creating text buffer...")
	tb := createEmptyTextBuffer()

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Insert content in chunks directly from file
	const insertChunkSize = 10 * 1024 * 1024 // 10MB chunks
	buffer := make([]byte, insertChunkSize)

	insertStart := time.Now()
	totalBytesRead := 0

	for {
		bytesRead, err := file.Read(buffer)
		if bytesRead > 0 {
			tb.Insert(totalBytesRead, string(buffer[:bytesRead]), true)
			totalBytesRead += bytesRead

			if (totalBytesRead/(1024*1024))%20 == 0 { // Report every 20MB
				fmt.Printf("Inserted %d/%d MB\n", totalBytesRead/(1024*1024), sizeMB)
				printMemUsage()
			}
		}

		if err != nil {
			break
		}
	}

	insertDuration := time.Since(insertStart)
	fmt.Printf("Total insertion of %dMB took: %v\n", sizeMB, insertDuration)
	printMemUsage()

	// Verify buffer size
	bufferLength := tb.GetLength()
	fmt.Printf("Buffer length: %d bytes\n", bufferLength)
	expectedSize := int64(sizeMB) * 1024 * 1024
	if int64(bufferLength) != expectedSize {
		t.Errorf("Buffer length mismatch: expected %d, got %d", expectedSize, bufferLength)
	}

	// Measure random access time
	fmt.Println("Performing random access operations...")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	accessStart := time.Now()

	for j := 0; j < 1000; j++ {
		offset := r.Intn(tb.GetLength())
		tb.NodeAt(offset)
	}

	accessDuration := time.Since(accessStart)
	fmt.Printf("1000 random accesses took: %v\n", accessDuration)

	// Measure line count
	lineCountStart := time.Now()
	lineCount := tb.GetLineCount()
	lineCountDuration := time.Since(lineCountStart)
	fmt.Printf("Line count (%d lines) took: %v\n", lineCount, lineCountDuration)

	// Measure random line retrievals
	fmt.Println("Performing random line retrievals...")
	lineStart := time.Now()

	for j := 0; j < 100; j++ {
		lineNum := r.Intn(lineCount) + 1 // Lines are 1-indexed
		tb.GetLineContent(lineNum)
	}

	lineDuration := time.Since(lineStart)
	fmt.Printf("100 random line retrievals took: %v\n", lineDuration)

	// Measure random small edits
	fmt.Println("Performing random edits...")
	editStart := time.Now()

	for j := 0; j < 100; j++ {
		offset := r.Intn(tb.GetLength())
		length := r.Intn(10) + 1 // Delete 1-10 characters
		if offset+length > tb.GetLength() {
			length = tb.GetLength() - offset
		}
		if length > 0 {
			tb.Delete(offset, length)
			tb.Insert(offset, "replacement", true)
		}
	}

	editDuration := time.Since(editStart)
	fmt.Printf("100 random edits took: %v\n", editDuration)

	// Summary
	fmt.Println("\nPerformance Summary (Medium File):")
	fmt.Printf("File Size: %d MB\n", sizeMB)
	fmt.Printf("Insert Time: %v\n", insertDuration)
	fmt.Printf("Random Access Time (1000 ops): %v\n", accessDuration)
	fmt.Printf("Line Count Time: %v\n", lineCountDuration)
	fmt.Printf("Line Retrieval Time (100 ops): %v\n", lineDuration)
	fmt.Printf("Edit Time (100 ops): %v\n", editDuration)
}
