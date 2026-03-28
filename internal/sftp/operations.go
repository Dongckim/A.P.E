package sftp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileInfo represents metadata about a remote file or directory.
type FileInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	IsDir       bool   `json:"is_dir"`
	Permissions string `json:"permissions"`
	ModTime     string `json:"mod_time"`
}

// ListDirectory returns the contents of a remote directory.
func (c *Client) ListDirectory(path string) ([]FileInfo, error) {
	entries, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %w", path, err)
	}

	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		files = append(files, FileInfo{
			Name:        entry.Name(),
			Path:        filepath.Join(path, entry.Name()),
			Size:        entry.Size(),
			IsDir:       entry.IsDir(),
			Permissions: entry.Mode().String(),
			ModTime:     entry.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	return files, nil
}

// ReadFile returns the full content of a remote file.
func (c *Client) ReadFile(path string) ([]byte, error) {
	f, err := c.sftpClient.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return data, nil
}

// WriteFile creates or overwrites a remote file with the given content.
func (c *Client) WriteFile(path string, content []byte) error {
	f, err := c.sftpClient.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

// UploadFile streams data from a reader into a remote file.
func (c *Client) UploadFile(path string, reader io.Reader) error {
	f, err := c.sftpClient.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("failed to upload file %s: %w", path, err)
	}
	return nil
}

// DownloadFile streams a remote file into the provided writer.
func (c *Client) DownloadFile(path string, w io.Writer) error {
	f, err := c.sftpClient.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("failed to download file %s: %w", path, err)
	}
	return nil
}

// DeleteFile removes a file or directory (recursive) at the given path.
func (c *Client) DeleteFile(path string) error {
	info, err := c.sftpClient.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if !info.IsDir() {
		if err := c.sftpClient.Remove(path); err != nil {
			return fmt.Errorf("failed to delete file %s: %w", path, err)
		}
		return nil
	}

	return c.removeDir(path)
}

// removeDir recursively removes a directory and all its contents.
func (c *Client) removeDir(path string) error {
	entries, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if err := c.removeDir(fullPath); err != nil {
				return err
			}
		} else {
			if err := c.sftpClient.Remove(fullPath); err != nil {
				return fmt.Errorf("failed to delete %s: %w", fullPath, err)
			}
		}
	}

	if err := c.sftpClient.RemoveDirectory(path); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}
	return nil
}

// RenameFile moves/renames a file or directory.
func (c *Client) RenameFile(oldPath, newPath string) error {
	if err := c.sftpClient.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", oldPath, newPath, err)
	}
	return nil
}

// Stat returns metadata about a single file or directory.
func (c *Client) Stat(path string) (*FileInfo, error) {
	info, err := c.sftpClient.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}
	return &FileInfo{
		Name:        info.Name(),
		Path:        path,
		Size:        info.Size(),
		IsDir:       info.IsDir(),
		Permissions: info.Mode().String(),
		ModTime:     info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}

// MkdirAll creates a directory and all parent directories.
func (c *Client) MkdirAll(path string) error {
	if err := c.sftpClient.MkdirAll(path); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}
