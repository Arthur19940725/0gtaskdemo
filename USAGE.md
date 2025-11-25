# 使用示例

## 快速开始

### 1. 准备测试文件

首先，您需要一个4GB的测试文件。可以使用以下命令创建一个测试文件（Linux/Mac）：

```bash
# 创建一个4GB的测试文件（使用随机数据）
dd if=/dev/urandom of=test_4gb.dat bs=1M count=4096

# 或者使用零填充（更快）
dd if=/dev/zero of=test_4gb.dat bs=1M count=4096
```

Windows PowerShell:
```powershell
# 创建一个4GB的测试文件
$file = [System.IO.File]::Create("test_4gb.dat")
$file.SetLength(4GB)
$file.Close()
```

### 2. 配置0g-storage-client

确保已安装并配置好0g-storage-client：

```bash
# 检查0g-storage-client是否安装
0g-storage-client --version

# 配置访问凭证（根据0g-storage-client文档）
0g-storage-client config
```

### 3. 运行程序

#### 方式一：完整流程（推荐）

```bash
go run main.go all test_4gb.dat merged_test_4gb.dat
```

这将自动执行：
1. 切分文件 → `./chunks/chunk_0.dat` 到 `chunk_9.dat`
2. 上传所有分块到0g-storage
3. 下载所有分块到 `./downloaded_chunks/`
4. 合并文件 → `merged_test_4gb.dat`

#### 方式二：分步执行

```bash
# 步骤1: 切分文件
go run main.go split test_4gb.dat

# 步骤2: 上传分块（需要配置好0g-storage-client）
go run main.go upload ./chunks

# 步骤3: 下载分块
go run main.go download ./downloaded_chunks

# 步骤4: 合并文件
go run main.go merge ./downloaded_chunks merged_test_4gb.dat
```

## 验证结果

### 验证文件切分

```bash
# 检查分块文件数量和大小
ls -lh ./chunks/

# 应该看到10个文件，每个约400MB
# chunk_0.dat, chunk_1.dat, ..., chunk_9.dat
```

### 验证文件合并

```bash
# 比较原始文件和合并后的文件大小
# Linux/Mac
ls -lh test_4gb.dat merged_test_4gb.dat

# Windows
dir test_4gb.dat merged_test_4gb.dat

# 使用文件校验和验证完整性（如果支持）
# Linux/Mac
md5sum test_4gb.dat merged_test_4gb.dat
# 或
sha256sum test_4gb.dat merged_test_4gb.dat
```

## 参数调整

如果需要修改分块大小或数量，编辑 `main.go` 中的常量：

```go
const (
    chunkSizeMB     = 400    // 修改每个分块的大小（MB）
    totalChunks     = 10     // 修改分块数量
    fragmentSizeMB  = 400    // 修改fragment size参数（MB）
)
```

例如，如果要切分成20个200MB的文件：

```go
const (
    chunkSizeMB     = 200
    totalChunks     = 20
    fragmentSizeMB  = 200
)
```

## 故障排除

### 问题1: 找不到0g-storage-client命令

**解决方案：**
- 确保0g-storage-client已安装并在PATH中
- 或者修改代码使用Go SDK方式

### 问题2: 上传/下载失败

**可能原因：**
- 0g-storage-client未正确配置访问凭证
- 网络连接问题
- fragment size参数不正确

**解决方案：**
- 检查0g-storage-client配置：`0g-storage-client config`
- 检查网络连接
- 根据0g-storage-client文档调整命令参数

### 问题3: 文件大小不是正好4GB

**说明：**
程序会自动处理文件大小不是正好4GB的情况。最后一个分块可能小于400MB。

### 问题4: 磁盘空间不足

**解决方案：**
- 确保有足够的磁盘空间（至少需要4GB+用于切分文件）
- 上传完成后可以删除 `./chunks/` 目录
- 下载和合并完成后可以删除 `./downloaded_chunks/` 目录

## 性能优化建议

1. **并发上传/下载**：可以修改代码支持并发处理多个分块
2. **断点续传**：对于大文件，可以实现断点续传功能
3. **进度显示**：可以添加进度条显示上传/下载进度
4. **错误重试**：可以添加自动重试机制

## 注意事项

1. 确保有足够的磁盘空间
2. 上传和下载需要稳定的网络连接
3. 大文件操作可能需要较长时间，请耐心等待
4. 建议在生产环境中添加日志记录和监控

