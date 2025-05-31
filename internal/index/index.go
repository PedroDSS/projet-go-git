package index

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type IndexEntry struct {
	Hash     string
	Filename string
}

func addSingleFile(filename string, indexEntries map[string]string) error {
	if strings.HasPrefix(filename, ".goit") {
		return nil
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("error accessing file %s: %v", filename, err)
	}

	if fileInfo.IsDir() {
		return nil
	}

	hash, content, err := processFile(filename)
	if err != nil {
		return err
	}

	objectPath := filepath.Join(".goit", "objects", hash)
	if err := os.WriteFile(objectPath, content, 0644); err != nil {
		return fmt.Errorf("error storing object %s: %v", objectPath, err)
	}

	indexEntries[filename] = hash
	return nil
}

func processFile(filename string) (string, []byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", nil, fmt.Errorf("error opening file %s: %v", filename, err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", nil, fmt.Errorf("error calculating hash for %s: %v", filename, err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	if _, err := f.Seek(0, 0); err != nil {
		return "", nil, fmt.Errorf("error seeking in file %s: %v", filename, err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return "", nil, fmt.Errorf("error reading content of %s: %v", filename, err)
	}

	return hash, content, nil
}

func loadIndexEntries() (map[string]string, error) {
	indexEntries := make(map[string]string)
	indexPath := filepath.Join(".goit", "index")

	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return indexEntries, nil
		}
		return nil, fmt.Errorf("error reading index: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(indexContent)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			hash, file := parts[0], parts[1]
			indexEntries[file] = hash
		}
	}

	return indexEntries, nil
}

func writeIndexEntries(indexEntries map[string]string) error {
	var lines []string
	for file, hash := range indexEntries {
		lines = append(lines, fmt.Sprintf("%s %s", hash, file))
	}

	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	indexPath := filepath.Join(".goit", "index")
	return os.WriteFile(indexPath, []byte(content), 0644)
}

func Add(filename string) {
	indexEntries, err := loadIndexEntries()
	if err != nil {
		fmt.Printf("Error loading index: %v\n", err)
		return
	}

	if filename == "." {
		addedCount := 0
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if path == ".goit" || strings.HasPrefix(path, ".goit"+string(os.PathSeparator)) {
					return filepath.SkipDir
				}
				return nil
			}

			if err := addSingleFile(path, indexEntries); err != nil {
				fmt.Printf("Warning: %v\n", err)
			} else {
				fmt.Printf("Added %s\n", path)
				addedCount++
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Error walking directory tree: %v\n", err)
			return
		}

		if addedCount == 0 {
			fmt.Println("No files were added")
			return
		}
	} else {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Printf("pathspec '%s' did not match any files\n", filename)
			return
		}

		if err := addSingleFile(filename, indexEntries); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Added %s\n", filename)
	}

	if err := writeIndexEntries(indexEntries); err != nil {
		fmt.Printf("Error writing index: %v\n", err)
	}
}

func GetIndexEntries() ([]IndexEntry, error) {
	indexEntries, err := loadIndexEntries()
	if err != nil {
		return nil, err
	}

	var entries []IndexEntry
	for filename, hash := range indexEntries {
		entries = append(entries, IndexEntry{
			Hash:     hash,
			Filename: filename,
		})
	}

	return entries, nil
}
