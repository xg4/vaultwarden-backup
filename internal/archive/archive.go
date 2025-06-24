package archive

import (
	"fmt"
	"io"
	"os"

	"github.com/xg4/vaultwarden-backup/pkg/crypto"
	"github.com/xg4/vaultwarden-backup/pkg/targz"
)

// EncryptedBackup creates an encrypted tar.gz archive of the specified directory.
func EncryptedBackup(backupDir, password, archiveFile string) error {
	outFile, err := os.Create(archiveFile)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer outFile.Close()

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		defer pipeWriter.Close()
		if err := targz.Create(backupDir, pipeWriter); err != nil {
			pipeWriter.CloseWithError(fmt.Errorf("failed to create tar.gz archive: %w", err))
		}
	}()

	if err := crypto.EncryptStream(pipeReader, outFile, password); err != nil {
		return fmt.Errorf("failed to encrypt archive: %w", err)
	}

	return nil
}

// DecryptBackup decrypts and extracts an encrypted backup.
func DecryptBackup(archiveFile, password, extractDir string) error {
	inFile, err := os.Open(archiveFile)
	if err != nil {
		return fmt.Errorf("failed to open archive file: %w", err)
	}
	defer inFile.Close()

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		defer pipeWriter.Close()
		if err := crypto.DecryptStream(inFile, pipeWriter, password); err != nil {
			pipeWriter.CloseWithError(fmt.Errorf("failed to decrypt archive: %w", err))
		}
	}()

	if err := targz.Extract(pipeReader, extractDir); err != nil {
		return fmt.Errorf("failed to extract tar.gz archive: %w", err)
	}

	return nil
}
