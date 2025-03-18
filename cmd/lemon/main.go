package main

import (
	"fmt"

	"github.com/kebaren/textbuffer/pkg/buffer"
)

func main() {

	// 这里可以添加应用程序的启动代码
	ptbb := buffer.NewPieceTreeTextBufferBuilder()
	ptbb.AcceptChunk("abc\n")
	ptbb.AcceptChunk("def")
	ptf := ptbb.Finish(true)
	tt := ptf.Create(buffer.LF)

	fmt.Println(tt.GetLineCount())
	fmt.Println(tt.GetLength())
	fmt.Printf("0->%q\n", tt.GetLineContent(0))
	fmt.Printf("1->%q\n", tt.GetLineContent(1))
	fmt.Printf("2->%q\n", tt.GetLineContent(2))
	fmt.Printf("%q\n", tt.GetLinesRawContent())

	tt.Insert(1, "+", true)
	fmt.Println(tt.GetLinesContent())
	fmt.Println(tt.GetLineCount())
	fmt.Println(tt.GetLength())
	fmt.Printf("0->%q\n", tt.GetLineContent(0))
	fmt.Printf("1->%q\n", tt.GetLineContent(1))
	fmt.Printf("2->%q\n", tt.GetLineContent(2))
	fmt.Printf("%q\n", tt.GetLinesRawContent())

	tt.Delete(2, 1)
	fmt.Println(tt.GetLinesContent())
	fmt.Println(tt.GetLineCount())
	fmt.Println(tt.GetLength())
	fmt.Printf("0->%q\n", tt.GetLineContent(0))
	fmt.Printf("1->%q\n", tt.GetLineContent(1))
	fmt.Printf("2->%q\n", tt.GetLineContent(2))
	fmt.Printf("%q\n", tt.GetLinesRawContent())

}
