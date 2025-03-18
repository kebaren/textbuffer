# Lemon Text Buffer

一个高性能的文本缓冲区实现，基于片段树（Piece Tree）数据结构。

## 项目结构

标准的Go语言项目结构：

```
.
├── cmd/                   # 命令行应用程序
│   └── lemon/             # 主应用程序入口点
├── pkg/                   # 可导出的库代码包
│   ├── buffer/            # 核心文本缓冲区实现
│   └── common/            # 通用工具和数据结构
├── internal/              # 私有应用程序和库代码
├── api/                   # API协议定义文件
├── configs/               # 配置文件模板
├── scripts/               # 脚本和工具
├── test/                  # 额外的外部测试
├── docs/                  # 文档
├── examples/              # 示例代码
└── tools/                 # 项目所需的工具
```

## 主要功能

- 高性能文本缓冲区实现
- 基于片段树（Piece Tree）数据结构
- 支持高效的插入、删除和查找操作
- 适用于编辑器和文本处理工具

## 使用方法

导入包：

```go
import "github.com/kebaren/textbuffer/pkg/buffer"
```

创建和使用文本缓冲区：

```go
// 创建一个新的文本缓冲区
tb := buffer.NewPieceTree("Hello, World!")

// 进行文本操作
tb.Insert(7, "Beautiful ")
text := tb.GetText() // "Hello, Beautiful World!"
```

## 许可证

[MIT License](LICENSE) 