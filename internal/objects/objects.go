package objects

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func CreateTree() (string, error) {
	indexContent, err := os.ReadFile(".goit/index")
	if err != nil {
		return "", fmt.Errorf("failed to read index: %v", err)
	}

	treeContent := "tree\n" + string(indexContent)
	hash := sha1.Sum([]byte(treeContent))
	hashStr := fmt.Sprintf("%x", hash[:])

	objectPath := filepath.Join(".goit", "objects", hashStr)
	if err := os.WriteFile(objectPath, []byte(treeContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write tree object: %v", err)
	}

	return hashStr, nil
}

func CreateCommit(treeHash, message, parentHash string) (string, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	content := fmt.Sprintf("commit\ntree %s\ndate %s\n", treeHash, now)
	if parentHash != "" {
		content += fmt.Sprintf("parent %s\n", parentHash)
	}
	content += fmt.Sprintf("\n%s", message)

	hash := sha1.Sum([]byte(content))
	hashStr := fmt.Sprintf("%x", hash[:])

	objectPath := filepath.Join(".goit", "objects", hashStr)
	if err := os.WriteFile(objectPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write commit object: %v", err)
	}

	return hashStr, nil
}

func ReadObject(hash string) (string, error) {
	objectPath := filepath.Join(".goit", "objects", hash)
	data, err := os.ReadFile(objectPath)
	if err != nil {
		return "", fmt.Errorf("object %s not found: %v", hash[:8], err)
	}
	return string(data), nil
}
