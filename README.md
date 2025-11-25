# 0gtaskdemo

一个使用Golang编写的程序，用于将4GB的大文件切分成10个400MB的文件，并使用0g-storage-client进行上传和下载。

## 功能特性

- ✅ 文件切分：将4GB文件切分成10个400MB的分块文件
- ✅ 文件上传：使用0g-storage-client上传分块文件（支持fragment size参数）
- ✅ 文件下载：从0g-storage下载分块文件
- ✅ 文件合并：将下载的分块文件合并回原始文件
- ✅ 完整流程：一键执行切分->上传->下载->合并的完整流程

## 配置参数

程序中的关键参数（可在`main.go`中修改）：

- `chunkSizeMB = 400`：每个分块文件的大小（MB）
- `totalChunks = 10`：分块文件的总数量
- `fragmentSizeMB = 400`：0g-storage-client上传/下载时的fragment size参数（MB）

## 使用方法

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 准备0g-storage-client

确保已安装并配置好`0g-storage-client`命令行工具，或者使用0g-storage-client的Go SDK。

**方式一：使用命令行工具**
- 安装0g-storage-client命令行工具
- 确保`0g-storage-client`命令在PATH中可用
- 配置好访问凭证和存储桶信息

**方式二：使用Go SDK**
- 修改代码中的`uploadChunksWithSDK`和`downloadChunksWithSDK`函数
- 参考0g-storage-client的Go SDK文档进行实现

### 3. 运行程序

#### 方式一：分步执行

```bash
# 1. 切分文件（将4GB文件切分成10个400MB文件）
go run main.go split <input_file>

# 2. 上传分块文件到0g-storage
go run main.go upload ./chunks

# 3. 下载分块文件
go run main.go download ./downloaded_chunks

# 4. 合并分块文件
go run main.go merge ./downloaded_chunks <output_file>
```

#### 方式二：一键执行完整流程

```bash
go run main.go all <input_file> <output_file>
```

这将自动执行：
1. 切分文件 → `./chunks/` 目录
2. 上传分块文件
3. 下载分块文件 → `./downloaded_chunks/` 目录
4. 合并文件 → 输出文件

## 示例

假设有一个4GB的文件`largefile.dat`：

```bash
# 完整流程示例
go run main.go all largefile.dat merged_largefile.dat

# 或者分步执行
go run main.go split largefile.dat
go run main.go upload ./chunks
go run main.go download ./downloaded_chunks
go run main.go merge ./downloaded_chunks merged_largefile.dat
```

## 目录结构

```
0gtaskdemo/
├── main.go          # 主程序文件
├── go.mod           # Go模块文件
├── README.md        # 说明文档
├── chunks/          # 切分后的分块文件目录（自动创建）
└── downloaded_chunks/ # 下载的分块文件目录（自动创建）
```

## Fragment Size参数说明

程序在上传和下载时使用`fragment size`参数设置为400MB，这与分块文件大小一致。这个参数控制0g-storage-client在传输文件时的分片大小。

**注意**：实际的0g-storage-client命令参数可能因版本而异，请根据实际情况调整代码中的命令参数。

## 错误处理

- 程序包含基本的错误处理机制
- 如果某个分块上传/下载失败，程序会继续处理其他分块
- 合并时会跳过不存在的分块文件

## 注意事项

1. **磁盘空间**：确保有足够的磁盘空间存储切分后的文件（至少4GB+）
2. **0g-storage配置**：确保0g-storage-client已正确配置访问凭证
3. **文件大小**：程序假设输入文件大小约为4GB，如果文件大小不同，可能需要调整`totalChunks`参数
4. **网络连接**：上传和下载需要稳定的网络连接

## 扩展开发

如果需要使用0g-storage-client的Go SDK，可以修改以下函数：

- `uploadChunksWithSDK()`：实现基于SDK的上传功能
- `downloadChunksWithSDK()`：实现基于SDK的下载功能

参考代码中已包含示例结构。

## 许可证

MIT License
