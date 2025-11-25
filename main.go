package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// 文件大小常量
	chunkSizeMB     = 400                    // 每个分块400MB
	chunkSizeBytes  = chunkSizeMB * 1024 * 1024 // 400MB in bytes
	totalChunks     = 10                     // 总共10个分块
	fragmentSizeMB  = 400                    // fragment size参数，400MB
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  go run main.go split <input_file>          - 切分文件")
		fmt.Println("  go run main.go upload <chunks_dir>          - 上传分块文件")
		fmt.Println("  go run main.go download <chunks_dir>       - 下载分块文件")
		fmt.Println("  go run main.go merge <chunks_dir> <output> - 合并文件")
		fmt.Println("  go run main.go all <input_file> <output>    - 执行完整流程（切分->上传->下载->合并）")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "split":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请指定输入文件路径")
			os.Exit(1)
		}
		splitFile(os.Args[2])

	case "upload":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请指定分块文件目录")
			os.Exit(1)
		}
		uploadChunks(os.Args[2])

	case "download":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请指定下载目录")
			os.Exit(1)
		}
		downloadChunks(os.Args[2])

	case "merge":
		if len(os.Args) < 4 {
			fmt.Println("错误: 请指定分块文件目录和输出文件路径")
			os.Exit(1)
		}
		mergeChunks(os.Args[2], os.Args[3])

	case "all":
		if len(os.Args) < 4 {
			fmt.Println("错误: 请指定输入文件路径和输出文件路径")
			os.Exit(1)
		}
		runFullProcess(os.Args[2], os.Args[3])

	default:
		fmt.Printf("未知命令: %s\n", command)
		os.Exit(1)
	}
}

// splitFile 将大文件切分成多个400MB的文件
func splitFile(inputFile string) {
	fmt.Printf("开始切分文件: %s\n", inputFile)

	// 打开输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("错误: 无法打开文件 %s: %v\n", inputFile, err)
		os.Exit(1)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("错误: 无法获取文件信息: %v\n", err)
		os.Exit(1)
	}

	fileSize := fileInfo.Size()
	fmt.Printf("文件大小: %.2f GB\n", float64(fileSize)/(1024*1024*1024))

	// 创建输出目录
	chunksDir := "./chunks"
	if err := os.MkdirAll(chunksDir, os.ModePerm); err != nil {
		fmt.Printf("错误: 无法创建目录 %s: %v\n", chunksDir, err)
		os.Exit(1)
	}

	// 使用io.CopyN进行切分，更高效
	chunkNum := 0

	for chunkNum < totalChunks {
		chunkFileName := filepath.Join(chunksDir, fmt.Sprintf("chunk_%d.dat", chunkNum))
		chunkFile, err := os.Create(chunkFileName)
		if err != nil {
			fmt.Printf("错误: 无法创建分块文件 %s: %v\n", chunkFileName, err)
			os.Exit(1)
		}

		// 使用io.CopyN复制指定大小的数据
		bytesWritten, err := io.CopyN(chunkFile, file, chunkSizeBytes)
		if err != nil && err != io.EOF {
			chunkFile.Close()
			fmt.Printf("错误: 写入分块文件失败: %v\n", err)
			os.Exit(1)
		}

		chunkFile.Close()

		if bytesWritten == 0 {
			break
		}

		fmt.Printf("已创建分块 %d: %s (%.2f MB)\n", chunkNum, chunkFileName, float64(bytesWritten)/(1024*1024))
		chunkNum++

		// 如果已经读取完整个文件，退出
		if bytesWritten < chunkSizeBytes {
			break
		}
	}

	fmt.Printf("文件切分完成! 共创建 %d 个分块文件\n", chunkNum)
}

// uploadChunks 上传所有分块文件到0g-storage
func uploadChunks(chunksDir string) {
	fmt.Printf("开始上传分块文件: %s\n", chunksDir)

	// 检查0g-storage-client是否可用
	if !checkCommandExists("0g-storage-client") {
		fmt.Println("警告: 未找到 0g-storage-client 命令")
		fmt.Println("请确保已安装 0g-storage-client 并在 PATH 中")
		fmt.Println("尝试使用Go SDK方式...")
		uploadChunksWithSDK(chunksDir)
		return
	}

	// 遍历所有分块文件并上传
	for i := 0; i < totalChunks; i++ {
		chunkFileName := filepath.Join(chunksDir, fmt.Sprintf("chunk_%d.dat", i))
		
		// 检查文件是否存在
		if _, err := os.Stat(chunkFileName); os.IsNotExist(err) {
			fmt.Printf("跳过不存在的文件: %s\n", chunkFileName)
			continue
		}

		fmt.Printf("正在上传分块 %d: %s\n", i, chunkFileName)

		// 构建上传命令，设置fragment size参数
		// 注意: 实际的命令参数可能因0g-storage-client版本而异
		// 这里假设使用 --fragment-size 参数
		cmd := exec.Command("0g-storage-client", 
			"upload",
			"--fragment-size", fmt.Sprintf("%dMB", fragmentSizeMB),
			chunkFileName)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("错误: 上传分块 %d 失败: %v\n", i, err)
			// 继续上传其他分块
			continue
		}

		fmt.Printf("成功上传分块 %d\n", i)
	}

	fmt.Println("所有分块文件上传完成!")
}

// uploadChunksWithSDK 使用Go SDK上传（如果可用）
func uploadChunksWithSDK(chunksDir string) {
	fmt.Println("注意: 此功能需要0g-storage-client的Go SDK")
	fmt.Println("请参考0g-storage-client的Go SDK文档进行实现")
	fmt.Println("示例代码结构:")
	fmt.Println(`
	import "github.com/0glabs/0g-storage-client/sdk"
	
	client := sdk.NewClient(config)
	for i := 0; i < totalChunks; i++ {
		chunkFile := filepath.Join(chunksDir, fmt.Sprintf("chunk_%d.dat", i))
		err := client.Upload(chunkFile, &sdk.UploadOptions{
			FragmentSize: fragmentSizeMB * 1024 * 1024,
		})
		// 处理错误...
	}
	`)
}

// downloadChunks 从0g-storage下载所有分块文件
func downloadChunks(outputDir string) {
	fmt.Printf("开始下载分块文件到: %s\n", outputDir)

	// 创建输出目录
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("错误: 无法创建目录 %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	// 检查0g-storage-client是否可用
	if !checkCommandExists("0g-storage-client") {
		fmt.Println("警告: 未找到 0g-storage-client 命令")
		fmt.Println("请确保已安装 0g-storage-client 并在 PATH 中")
		fmt.Println("尝试使用Go SDK方式...")
		downloadChunksWithSDK(outputDir)
		return
	}

	// 遍历所有分块并下载
	for i := 0; i < totalChunks; i++ {
		chunkFileName := filepath.Join(outputDir, fmt.Sprintf("chunk_%d.dat", i))
		
		fmt.Printf("正在下载分块 %d: %s\n", i, chunkFileName)

		// 构建下载命令
		// 注意: 实际的命令参数可能因0g-storage-client版本而异
		// 这里假设需要指定文件ID或路径
		cmd := exec.Command("0g-storage-client",
			"download",
			"--fragment-size", fmt.Sprintf("%dMB", fragmentSizeMB),
			"--output", chunkFileName,
			fmt.Sprintf("chunk_%d.dat", i)) // 假设这是文件ID或路径

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("错误: 下载分块 %d 失败: %v\n", i, err)
			// 继续下载其他分块
			continue
		}

		fmt.Printf("成功下载分块 %d\n", i)
	}

	fmt.Println("所有分块文件下载完成!")
}

// downloadChunksWithSDK 使用Go SDK下载（如果可用）
func downloadChunksWithSDK(outputDir string) {
	fmt.Println("注意: 此功能需要0g-storage-client的Go SDK")
	fmt.Println("请参考0g-storage-client的Go SDK文档进行实现")
	fmt.Println("示例代码结构:")
	fmt.Println(`
	import "github.com/0glabs/0g-storage-client/sdk"
	
	client := sdk.NewClient(config)
	for i := 0; i < totalChunks; i++ {
		chunkFile := filepath.Join(outputDir, fmt.Sprintf("chunk_%d.dat", i))
		err := client.Download(fmt.Sprintf("chunk_%d.dat", i), chunkFile, &sdk.DownloadOptions{
			FragmentSize: fragmentSizeMB * 1024 * 1024,
		})
		// 处理错误...
	}
	`)
}

// mergeChunks 合并所有分块文件
func mergeChunks(chunksDir string, outputFile string) {
	fmt.Printf("开始合并分块文件: %s -> %s\n", chunksDir, outputFile)

	// 创建输出文件
	mergedFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("错误: 无法创建输出文件 %s: %v\n", outputFile, err)
		os.Exit(1)
	}
	defer mergedFile.Close()

	// 按顺序合并所有分块
	for i := 0; i < totalChunks; i++ {
		chunkFileName := filepath.Join(chunksDir, fmt.Sprintf("chunk_%d.dat", i))
		
		// 检查文件是否存在
		if _, err := os.Stat(chunkFileName); os.IsNotExist(err) {
			fmt.Printf("警告: 分块文件不存在: %s，跳过\n", chunkFileName)
			continue
		}

		chunkFile, err := os.Open(chunkFileName)
		if err != nil {
			fmt.Printf("错误: 无法打开分块文件 %s: %v\n", chunkFileName, err)
			continue
		}

		// 复制分块内容到合并文件
		bytesCopied, err := io.Copy(mergedFile, chunkFile)
		if err != nil {
			chunkFile.Close()
			fmt.Printf("错误: 合并分块 %d 失败: %v\n", i, err)
			continue
		}

		chunkFile.Close()
		fmt.Printf("已合并分块 %d: %.2f MB\n", i, float64(bytesCopied)/(1024*1024))
	}

	// 获取合并后文件的大小
	mergedInfo, _ := mergedFile.Stat()
	fmt.Printf("文件合并完成! 输出文件: %s (%.2f GB)\n", 
		outputFile, float64(mergedInfo.Size())/(1024*1024*1024))
}

// runFullProcess 执行完整流程
func runFullProcess(inputFile string, outputFile string) {
	fmt.Println("========== 开始完整流程 ==========")
	
	chunksDir := "./chunks"
	downloadDir := "./downloaded_chunks"

	// 1. 切分文件
	fmt.Println("\n[步骤 1/4] 切分文件")
	splitFile(inputFile)

	// 2. 上传分块
	fmt.Println("\n[步骤 2/4] 上传分块文件")
	uploadChunks(chunksDir)

	// 3. 下载分块
	fmt.Println("\n[步骤 3/4] 下载分块文件")
	// 创建下载目录
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		fmt.Printf("错误: 无法创建下载目录: %v\n", err)
		os.Exit(1)
	}
	downloadChunks(downloadDir)

	// 4. 合并文件
	fmt.Println("\n[步骤 4/4] 合并文件")
	mergeChunks(downloadDir, outputFile)

	fmt.Println("\n========== 完整流程执行完成! ==========")
}

// checkCommandExists 检查命令是否存在
func checkCommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

