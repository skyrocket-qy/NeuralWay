package tools

import (
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AssetCompressor compresses game assets for web deployment.
type AssetCompressor struct {
	SourceDir string
	OutputDir string
}

// NewAssetCompressor creates a new compressor.
func NewAssetCompressor(src, out string) *AssetCompressor {
	return &AssetCompressor{
		SourceDir: src,
		OutputDir: out,
	}
}

// CompressAll processes all assets in the source directory.
func (c *AssetCompressor) CompressAll() error {
	if err := os.MkdirAll(c.OutputDir, 0o755); err != nil {
		return err
	}

	return filepath.Walk(c.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(c.SourceDir, path)
		if err != nil {
			return err
		}

		outPath := filepath.Join(c.OutputDir, relPath)

		// Create output directory structure
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}

		// Compress based on file type
		if shouldGzip(path) {
			return gzipFile(path, outPath+".gz")
		} else {
			// Just copy non-compressible files
			return copyFile(path, outPath)
		}
	})
}

// CreateAssetPack creates a single ZIP file containing all assets.
func (c *AssetCompressor) CreateAssetPack(dest string) error {
	zipFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	return filepath.Walk(c.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(c.SourceDir, path)
		if err != nil {
			return err
		}

		// Create zip file entry
		writer, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		// Read source file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Write to zip
		_, err = io.Copy(writer, file)

		return err
	})
}

func shouldGzip(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json", ".txt", ".xml", ".csv", ".shader", ".js", ".wasm":
		return true
	default:
		return false // Images/Audio usually already compressed
	}
}

func gzipFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	writer := gzip.NewWriter(destination)
	defer writer.Close()

	_, err = io.Copy(writer, source)

	return err
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)

	return err
}
