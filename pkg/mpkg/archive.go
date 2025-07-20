package mpkg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
	"github.com/sirupsen/logrus"
)

var compressions = map[string]archives.Compression{
	"gz":   archives.Gz{},
	"bz2":  archives.Bz2{},
	"xz":   archives.Xz{},
	"zst":  archives.Zstd{},
	"lz4":  archives.Lz4{},
	"br":   archives.Brotli{},
	"lzip": archives.Lzip{},
	"sz":   archives.Sz{},
	"zlib": archives.Zlib{},
}

var archivals = map[string]archives.Archival{
	"tar": archives.Tar{},
	"zip": archives.Zip{},
}

func ArchiveFiles(ctx context.Context, dir string, dst string, filenames map[string]string, archival string, compression string) error {
	com := compressions[compression]
	if com == nil {
		return fmt.Errorf("unsupported compression %s", compression)
	}
	archiv := archivals[archival]
	if archiv == nil {
		return fmt.Errorf("unsupported archival %s", archival)
	}

	if IsNotExist(dir) {
		return fmt.Errorf("no such directory %s", dir)
	}
	if !IsNotExist(dst) {
		return fmt.Errorf("%s already exists", dst)
	}

	files, err := archives.FilesFromDisk(ctx, &archives.FromDiskOptions{FollowSymlinks: true}, filenames)
	if err != nil {
		return fmt.Errorf("unable to map files to directory: %w", err)
	}

	logrus.Infof("Creating package file %s", dst)
	dstf, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("unable to create a new %s file: %w", dst, err)
	}
	defer dstf.Close()
	format := archives.CompressedArchive{
		Compression: com,
		Archival:    archiv,
	}

	err = format.Archive(ctx, dstf, files)
	if err != nil {
		return fmt.Errorf("unable to archive: %w", err)
	}

	return nil
}

func Archive(ctx context.Context, dir string, dst string, archival string, compression string) error {
	baseDirName := filepath.Base(filepath.Clean(dir))
	if dir == "." {
		baseDirName = ""
	}
	return ArchiveFiles(ctx, dir, dst, map[string]string{dir: baseDirName}, compression, archival)
}

// Unarchive unpacks the given compressed file to destination
func Unarchive(ctx context.Context, filepath string, dest string) error {
	archivef, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error while opening file: %w", err)
	}
	defer archivef.Close()

	format, input, err := archives.Identify(ctx, filepath, archivef)
	if err != nil {
		return err
	}
	extractor, ok := format.(archives.Extractor)
	if !ok {
		return errors.New("unsupported archive format for extraction")
	}

	handler := func(ctx context.Context, f archives.FileInfo) error {
		return handleArchivedFile(f, dest)
	}
	err = extractor.Extract(ctx, input, handler)
	if err != nil {
		return fmt.Errorf("error while extracting files: %w", err)
	}
	return nil
}

// Borrowed from https://github.com/jm33-m0/arc/blob/main/v2/unarchiver.go
func handleArchivedFile(file archives.FileInfo, dest string) error {
	dstPath, err := SecurePath(dest, file.NameInArchive)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(dstPath)
	if err := CreateDirWithPerm(parentDir, 0o700); err != nil {
		return fmt.Errorf("mkdir %s: %w", parentDir, err)
	}

	if file.IsDir() {
		return CreateDirWithPerm(dstPath, file.Mode())
	}

	// TODO: Handle symlinks without dereferencing it
	if file.LinkTarget != "" {
		return nil
	}

	originMode, err := os.Stat(parentDir)
	if err != nil {
		return fmt.Errorf("unable to stat %s: %w", parentDir, err)
	}
	if originMode.Mode().Perm()&0o200 == 0 {
		if err := os.Chmod(parentDir, originMode.Mode()|0o200); err != nil {
			return fmt.Errorf("chmod parent directory: %w", err)
		}
		defer func() {
			_ = os.Chmod(parentDir, originMode.Mode())
		}()
	}

	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, file.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, reader); err != nil {
		return fmt.Errorf("error while copying archive file: %w", err)
	}

	return nil
}

func SecurePath(basePath string, relativePath string) (string, error) {
	relativePath = filepath.Clean("/" + relativePath)
	relativePath = strings.TrimPrefix(relativePath, string(os.PathSeparator))

	dstPath := filepath.Join(basePath, relativePath)
	securedPath := filepath.Clean(dstPath) + string(os.PathSeparator)
	if basePath != "/" {
		basePath = filepath.Clean(basePath) + string(os.PathSeparator)
	}

	if !strings.HasPrefix(securedPath, basePath) {
		return "", fmt.Errorf("illegal file path: %s", dstPath)
	}

	return dstPath, nil
}
