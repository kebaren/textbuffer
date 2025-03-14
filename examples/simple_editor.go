package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kebaren/lemon/pkg/buffer"
)

func main() {
	// 创建一个新的文本缓冲区
	initialText := "Welcome to Lemon Text Buffer example!\n" +
		"Type commands to interact with the buffer:\n" +
		"- insert <pos> <text>: Insert text at position\n" +
		"- delete <start> <end>: Delete text from start to end\n" +
		"- print: Print current buffer content\n" +
		"- exit: Exit the program\n"

	tree := buffer.NewPieceTree(initialText)

	fmt.Println(initialText)

	// 简单的命令行交互
	for {
		fmt.Print("> ")
		var line string
		fmt.Scanln(&line)

		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "insert":
			if len(args) < 3 {
				fmt.Println("Usage: insert <pos> <text>")
				continue
			}
			var pos int
			fmt.Sscanf(args[1], "%d", &pos)
			text := strings.Join(args[2:], " ")
			tree.Insert(pos, text, false)
			fmt.Println("Text inserted.")

		case "delete":
			if len(args) < 3 {
				fmt.Println("Usage: delete <start> <end>")
				continue
			}
			var start, end int
			fmt.Sscanf(args[1], "%d", &start)
			fmt.Sscanf(args[2], "%d", &end)
			tree.Delete(start, end-start)
			fmt.Println("Text deleted.")

		case "print":
			fmt.Println(tree.GetLinesContent())

		case "exit":
			os.Exit(0)

		default:
			fmt.Println("Unknown command.")
		}
	}
}
