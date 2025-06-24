package main

import (
	"fmt"
	"os"

	"github.com/xg4/vaultwarden-backup/internal/archive"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input_directory> <output_encrypted_file> <password>\n", os.Args[0])
		os.Exit(1)
	}

	inputDir := os.Args[1]
	outEncryptedFile := os.Args[2]
	password := os.Args[3]

	if err := archive.EncryptedBackup(inputDir, password, outEncryptedFile); err != nil {
		fmt.Fprintf(os.Stderr, "创建加密归档失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(" Done.")
	fmt.Println("Backup complete.")
}
