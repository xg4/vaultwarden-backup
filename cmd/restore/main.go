package main

import (
	"fmt"
	"os"

	"github.com/xg4/vaultwarden-backup/internal/archive"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input_encrypted_file> <output_directory> <password>\n", os.Args[0])
		os.Exit(1)
	}

	inputEncryptedFile := os.Args[1]
	outputDir := os.Args[2]
	password := os.Args[3]

	if err := archive.DecryptBackup(inputEncryptedFile, password, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "解密归档失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(" Done.")
	fmt.Println("Restore complete.")
}
