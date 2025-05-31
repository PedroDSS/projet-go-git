package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/objects"
	"strings"
)

func IsGoitRepo() bool {
	_, err := os.Stat(".goit")
	return err == nil
}

func Init() {
	if IsGoitRepo() {
		fmt.Println("Already a goit repository")
		return
	}

	dirs := []string{
		".goit",
		".goit/objects",
		".goit/refs",
		".goit/refs/heads",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", dir, err)
			return
		}
	}

	// Initialise le HEAD pour pointer vers la branche master
	headContent := "ref: refs/heads/master"
	if err := os.WriteFile(".goit/HEAD", []byte(headContent), 0644); err != nil {
		fmt.Printf("Failed to create HEAD: %v\n", err)
		return
	}

	fmt.Println("Initialized empty Goit repository in .goit/")
}

func Commit(message string) {
	indexPath := filepath.Join(".goit", "index")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Println("Nothing to commit (create/copy files and use \"goit add\" to track)")
		return
	}

	treeHash, err := objects.CreateTree()
	if err != nil {
		fmt.Printf("Failed to create tree: %v\n", err)
		return
	}

	// Get parent commit if exists
	var parentHash string
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		fmt.Printf("Error getting current branch: %v\n", err)
		return
	}

	branchPath := filepath.Join(".goit", "refs", "heads", currentBranch)
	if data, err := os.ReadFile(branchPath); err == nil {
		parentHash = strings.TrimSpace(string(data))
	}

	commitHash, err := objects.CreateCommit(treeHash, message, parentHash)
	if err != nil {
		fmt.Printf("Failed to create commit: %v\n", err)
		return
	}

	if err := os.WriteFile(branchPath, []byte(commitHash), 0644); err != nil {
		fmt.Printf("Failed to update branch reference: %v\n", err)
		return
	}

	// Nettoyer l'index après le commit.
	os.Remove(indexPath)

	fmt.Printf("Committed: %s\n", commitHash[:8])
}

func GetCurrentBranch() (string, error) {
	head, err := GetHEAD()
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(head, "ref: refs/heads/") {
		return strings.TrimPrefix(head, "ref: refs/heads/"), nil
	}

	return "HEAD", nil // Etat HEAD détaché
}

func SetHEAD(ref string) error {
	return os.WriteFile(".goit/HEAD", []byte(ref), 0644)
}

func GetHEAD() (string, error) {
	data, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD: %v", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func GetCurrentCommitHash() (string, error) {
	head, err := GetHEAD()
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(head, "ref: ") {
		// HEAD pointe sur une branch
		refPath := strings.TrimPrefix(head, "ref: ")
		refFile := filepath.Join(".goit", refPath)
		data, err := os.ReadFile(refFile)
		if err != nil {
			return "", nil // Pas de commits pour l'instant
		}
		return strings.TrimSpace(string(data)), nil
	}
	// HEAD pointe directement sur un commit (HEAD détaché)
	return head, nil
}
