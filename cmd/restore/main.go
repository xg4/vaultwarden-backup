package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/archive"
)

var (
	inputFile = flag.String("input", "", "输入的加密备份文件路径 (必需)")
	outputDir = flag.String("output", "", "输出目录路径 (必需)")
	password  = flag.String("password", "", "解密密码 (必需)")
	verbose   = flag.Bool("verbose", false, "启用详细输出")
	help      = flag.Bool("help", false, "显示帮助信息")
)

func init() {
	// 添加简写选项
	flag.StringVar(inputFile, "i", "", "输入的加密备份文件路径 (必需)")
	flag.StringVar(outputDir, "o", "", "输出目录路径 (必需)")
	flag.StringVar(password, "p", "", "解密密码 (必需)")
	flag.BoolVar(verbose, "v", false, "启用详细输出")
	flag.BoolVar(help, "h", false, "显示帮助信息")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Vaultwarden 备份解密工具\n\n")
	fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "选项:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n示例:\n")
	fmt.Fprintf(os.Stderr, "  %s -i backup.enc -o ./restored -p mypassword\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s -input backup.enc -output ./restored -password mypassword -verbose\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "  %s -i backup.enc -o ./restored -p mypassword -v\n", filepath.Base(os.Args[0]))
}

func validateArgs() error {
	if *inputFile == "" {
		return fmt.Errorf("必须指定输入文件 (-input)")
	}

	if *outputDir == "" {
		return fmt.Errorf("必须指定输出目录 (-output)")
	}

	if *password == "" {
		return fmt.Errorf("必须指定解密密码 (-password)")
	}

	// 检查输入文件是否存在
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		return fmt.Errorf("输入文件不存在: %s", *inputFile)
	}

	return nil
}

func main() {
	// 自定义 usage 函数
	flag.Usage = usage

	// 解析命令行参数
	flag.Parse()

	// 如果指定了 help 标志，显示帮助并退出
	if *help {
		usage()
		os.Exit(0)
	}

	// 验证参数
	if err := validateArgs(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n\n", err)
		usage()
		os.Exit(1)
	}

	// 详细输出模式
	if *verbose {
		fmt.Printf("输入文件: %s\n", *inputFile)
		fmt.Printf("输出目录: %s\n", *outputDir)
		fmt.Printf("开始解密...\n")
	}

	// 执行解密
	if err := archive.DecryptBackup(*inputFile, *password, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "解密归档失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done.")
	fmt.Println("Restore complete.")

	if *verbose {
		fmt.Printf("文件已成功解密到: %s\n", *outputDir)
	}
}
