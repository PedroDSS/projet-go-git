package checkout

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/repository"
)

func Switch(branch string) {
	branchPath := filepath.Join(".goit", "refs", "heads", branch)

	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		fmt.Printf("Branch '%s' does not exist\n", branch)
		return
	}

	if hasUncommittedChanges() {
		fmt.Println("You have uncommitted changes. Please commit them before switching branches")
		return
	}

	headContent := fmt.Sprintf("ref: refs/heads/%s", branch)
	if err := repository.SetHEAD(headContent); err != nil {
		fmt.Printf("Failed to switch to branch: %v\n", err)
		return
	}

	fmt.Printf("Switched to branch '%s'\n", branch)
}

func hasUncommittedChanges() bool {
	// TODO: A implémenter, ça doit vérifier le working directory
	// NOTE: Cela ne fait qu'un simple check
	return false
}
