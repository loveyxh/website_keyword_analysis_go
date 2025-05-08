# 网站关键词检测工具

## 项目功能描述

这是一个用 Go 语言开发的网站数据分析工具，主要用于批量检测网站源代码中是否包含特定关键词。该工具具有以下功能：

- 从 Excel 文件中读取网站列表数据
- 并发获取多个网站的源代码内容
- 检查源代码中是否包含预设关键词（默认为"关键词1"和"关键词2"）
- 将检测结果导出到新的 Excel 文件中，包括匹配结果和错误信息
- 提供详细的运行日志和统计分析

## 项目安装

### 方法一：使用可执行文件（推荐）

本项目提供了打包好的可执行文件 `website_analysis.exe`，无需安装任何依赖即可直接运行。

### 方法二：从源码安装

如果您希望从源码安装和运行，请确保您的环境满足以下要求：

1. 安装 Go 语言环境（推荐 Go 1.18 或更高版本）
2. 克隆本仓库到本地

```bash
git clone https://github.com/loveyxh/website_keyword_analysis_go.git
cd website_analysis
```

3. 安装依赖包

依赖包包括：
- github.com/PuerkitoBio/goquery：用于HTML解析和处理
- github.com/xuri/excelize/v2：用于Excel文件的读写操作
- golang.org/x/net：网络通信包
- golang.org/x/text：文本处理包

```bash
go mod tidy
```

### 项目结构

```
website_analysis/
├── main.go              # 主程序入口
├── go.mod               # Go模块依赖定义
├── go.sum               # 依赖包校验和
├── run.bat              # Windows批处理启动脚本
├── website_analysis.exe # 预编译可执行文件
├── README.md            # 项目说明文档
└── 网站列表_样例数据.xlsx  # 示例输入文件
```

## 项目运行

### 方法一：使用可执行文件

1. 准备包含网站列表的 Excel 文件，确保文件中有 `media_url` 列
2. 双击 `run.bat` 文件运行程序，或直接运行以下命令：

```bash
website_analysis.exe
```

### 方法二：从源码运行

在项目目录下执行：

```bash
go run main.go
```

### 输入和输出

- 默认输入文件：`网站列表_样例数据.xlsx`（必须包含 `media_url` 列）
- 默认输出文件：`网站列表_样例数据_结果_[时间戳].xlsx`（包含新增的 `match_result` 和 `error_msg` 列）

### 结果说明

- `match_result = 1`：网站源代码中包含关键词
- `match_result = 0`：网站源代码中不包含关键词或获取失败
- `error_msg`：记录获取网站源代码时的错误信息（如网站无法访问）

### 高级配置

程序中有一些可调整的参数（需修改源码）：

- `maxConcurrent`：并发处理的网站数量（默认为 5）
- `requestTimeout`：请求超时时间（默认为 30 秒）
- `requestInterval`：请求间隔时间（默认为 1 秒）
- `keywords`：关键词列表（默认为 ["八项规定", "八项规定精神"]） 