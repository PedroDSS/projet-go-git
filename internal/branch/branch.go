package branch

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/repository"
	"strings"
)

func Create(name string) {
	if strings.Contains(name, "/") || strings.Contains(name, " ") {
		fmt.Printf("Invalid branch name: %s\n", name)
		return
	}

	branchPath := filepath.Join(".goit", "refs", "heads", name)

	if _, err := os.Stat(branchPath); err == nil {
		fmt.Printf("Branch '%s' already exists\n", name)
		return
	}

	currentHash, err := repository.GetCurrentCommitHash()
	if err != nil || currentHash == "" {
		fmt.Println("Cannot create branch: no commits yet")
		return
	}

	if err := os.WriteFile(branchPath, []byte(currentHash), 0644); err != nil {
		fmt.Printf("Failed to create branch: %v\n", err)
		return
	}

	fmt.Printf("Branch '%s' created\n", name)
}

func List() {
	branchesDir := filepath.Join(".goit", "refs", "heads")
	entries, err := os.ReadDir(branchesDir)
	if err != nil {
		fmt.Printf("Cannot list branches: %v\n", err)
		return
	}

	currentBranch, _ := repository.GetCurrentBranch()

	for _, entry := range entries {
		if !entry.IsDir() {
			prefix := "  "
			if entry.Name() == currentBranch {
				prefix = "* "
			}
			fmt.Printf("%s%s\n", prefix, entry.Name())
		}
	}
}
