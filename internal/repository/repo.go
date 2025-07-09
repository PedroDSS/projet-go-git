package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/objects"
	"strings"
)

func Init() {
	os.Mkdir(".goit", 0755)
	os.Mkdir(".goit/objects", 0755)
	os.Mkdir(".goit/refs", 0755)
	os.Mkdir(".goit/refs/heads", 0755)
	os.WriteFile(".goit/HEAD", []byte("ref: refs/heads/master"), 0644)
	fmt.Println("Initialized empty Goit repository")
}

func Commit(message string) {
	treeHash := objects.CreateTree()

	// Récupérer le hash du commit parent (HEAD actuel)
	var parentHash string
	headContent := GetHEAD()
	if strings.HasPrefix(headContent, "ref: ") {
		refPath := strings.TrimPrefix(headContent, "ref: ")
		refFile := filepath.Join(".goit", refPath)
		if content, err := os.ReadFile(refFile); err == nil {
			parentHash = strings.TrimSpace(string(content))
		}
	}

	commitHash := objects.CreateCommit(treeHash, message, parentHash)

	// Résoudre la référence HEAD actuelle
	headContent = GetHEAD()
	var refPath string
	if strings.HasPrefix(headContent, "ref: ") {
		refPath = strings.TrimPrefix(headContent, "ref: ")
	} else {
		// Si HEAD contient directement un hash, utiliser master par défaut
		refPath = "refs/heads/master"
	}

	// Écrire le hash du commit dans la référence
	os.WriteFile(filepath.Join(".goit", refPath), []byte(commitHash), 0644)

	// Vider l'index après le commit
	os.Remove(".goit/index")

	fmt.Println("Committed:", commitHash)
}

func SetHEAD(ref string) error {
	return os.WriteFile(".goit/HEAD", []byte(ref), 0644)
}

func GetHEAD() string {
	data, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
