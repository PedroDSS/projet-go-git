package status

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"projet-go-git/internal/index"
	"projet-go-git/internal/repository"
	"strings"
)

func ShowStatus() {
	fmt.Printf("On branch %s\n\n", getCurrentBranchName())

	entries, err := index.GetIndexEntries()
	var indexEntries map[string]string

	if err != nil {
		indexEntries = loadIndexDirect()
	} else {
		indexEntries = make(map[string]string)
		for _, entry := range entries {
			indexEntries[entry.Filename] = entry.Hash
		}
	}

	if len(indexEntries) > 0 {
		fmt.Println("Changes to be committed:")
		for filename := range indexEntries {
			fmt.Printf("  new file:   %s\n", filename)
		}
		fmt.Println()
	}

	var modified []string
	var untracked []string

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldIgnoreDirectory(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldIgnoreFile(path) {
			return nil
		}

		if storedHash, isTracked := indexEntries[path]; isTracked {
			if hasFileChanged(path, storedHash) {
				modified = append(modified, path)
			}
		} else {
			untracked = append(untracked, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	if len(modified) > 0 {
		fmt.Println("Changes not staged for commit:")
		for _, file := range modified {
			fmt.Printf("  modified:   %s\n", file)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println("Untracked files:")
		for _, file := range untracked {
			fmt.Printf("  %s\n", file)
		}
		fmt.Println()
		fmt.Println("Use 'goit add <file>' to include in what will be committed")
	}

	if len(indexEntries) == 0 && len(modified) == 0 && len(untracked) == 0 {
		fmt.Println("nothing to commit, working tree clean")
	}
}

func getCurrentBranchName() string {
	branch, err := repository.GetCurrentBranch()
	if err != nil {
		return "master" // nom de branche par défaut
	}
	return branch
}

func loadIndexDirect() map[string]string {
	indexEntries := make(map[string]string)
	indexContent, err := os.ReadFile(".goit/index")
	if err != nil {
		return indexEntries
	}

	entries := strings.Split(string(indexContent), "\n")
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, " ", 2)
		if len(parts) == 2 {
			hash, filename := parts[0], parts[1]
			indexEntries[filename] = hash
		}
	}
	return indexEntries
}

func shouldIgnoreDirectory(path string) bool {
	ignoredDirs := []string{".goit", ".git"}
	for _, ignored := range ignoredDirs {
		if path == ignored || strings.HasPrefix(path, ignored+"/") {
			return true
		}
	}
	return false
}

func shouldIgnoreFile(path string) bool {
	if strings.HasPrefix(path, ".goit") || strings.HasPrefix(path, ".git") {
		return true
	}

	ignoredFiles := []string{
		".gitignore",
		"goit",
		".DS_Store",
		"Thumbs.db",
		"desktop.ini",
	}

	for _, ignored := range ignoredFiles {
		if path == ignored {
			return true
		}
	}

	return false
}

func hasFileChanged(filename, indexHash string) bool {
	currentHash, err := hashFile(filename)
	if err != nil {
		return true // Fichier supprimé ou inaccessible
	}

	return currentHash != indexHash
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha1.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func ShowDiff(filename string) {
	if filename == "" {
		showAllDiffs()
		return
	}
	showFileDiff(filename)
}

func showAllDiffs() {
	indexEntries := loadIndexDirect()
	if len(indexEntries) == 0 {
		fmt.Println("No staged files to diff")
		return
	}

	hasChanges := false
	for filename, indexHash := range indexEntries {
		if hasFileChanged(filename, indexHash) {
			if !hasChanges {
				fmt.Println("Differences found:")
				hasChanges = true
			}
			fmt.Printf("\ndiff --goit a/%s b/%s\n", filename, filename)
			showFileDiff(filename)
		}
	}

	if !hasChanges {
		fmt.Println("No differences found")
	}
}

func showFileDiff(filename string) {
	fmt.Printf("File: %s\n", filename)

	indexEntries := loadIndexDirect()
	indexHash, exists := indexEntries[filename]
	if !exists {
		fmt.Println("File not staged")
		return
	}

	objectPath := filepath.Join(".goit", "objects", indexHash)
	stagedContent, err := os.ReadFile(objectPath)
	if err != nil {
		fmt.Printf("Cannot read staged version: %v\n", err)
		return
	}

	workingContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Working file deleted or inaccessible")
		return
	}

	if string(stagedContent) != string(workingContent) {
		fmt.Println("--- staged version")
		fmt.Println("+++ working version")
		fmt.Println("Files differ (detailed line-by-line diff not implemented)")

		fmt.Printf("Staged version: %d bytes\n", len(stagedContent))
		fmt.Printf("Working version: %d bytes\n", len(workingContent))
	} else {
		fmt.Println("No differences")
	}
}
