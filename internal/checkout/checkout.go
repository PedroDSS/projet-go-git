package checkout

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/repository"
)

func Switch(branch string) {
	refPath := filepath.Join(".goit", "refs", "heads", branch)
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		fmt.Println("Branch does not exist:", branch)
		return
	}

	hashBytes, err := os.ReadFile(refPath)
	if err != nil {
		fmt.Println("Unable to read branch ref:", err)
		return
	}

	hash := string(hashBytes)
	repository.SetHEAD(hash)
	fmt.Println("Switched to branch:", branch)
}