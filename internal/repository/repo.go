package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/objects"
	"strings"
)

/**
 * Vérifie si le répertoire actuel est un repository goit
 */
func IsGoitRepo() bool {
	_, err := os.Stat(".goit")
	return err == nil
}

/**
 * Initialise un nouveau repository goit
 * Crée la structure de dossiers et les fichiers de base
 */
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

	// Initialise le HEAD pour pointer vers la branche main
	headContent := "ref: refs/heads/main"
	if err := os.WriteFile(".goit/HEAD", []byte(headContent), 0644); err != nil {
		fmt.Printf("Failed to create HEAD: %v\n", err)
		return
	}

	// Crée un index vide
	if err := os.WriteFile(".goit/index", []byte(""), 0644); err != nil {
		fmt.Printf("Failed to create index: %v\n", err)
		return
	}

	fmt.Println("Initialized empty goit repository")
}

/**
 * Crée un nouveau commit avec les fichiers de l'index
 * Gère les commits parents et met à jour les références
 */
func Commit(message string) {
	indexPath := filepath.Join(".goit", "index")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Println("Nothing to commit (create/copy files and use \"goit add\" to track)")
		return
	}

	treeHash := objects.CreateTree()

	// Récupérer le hash du commit parent (HEAD actuel)
	var parentHash string
	headContent, err := GetHEAD()
	if err == nil && strings.HasPrefix(headContent, "ref: ") {
		refPath := strings.TrimPrefix(headContent, "ref: ")
		refFile := filepath.Join(".goit", refPath)
		if content, err := os.ReadFile(refFile); err == nil {
			parentHash = strings.TrimSpace(string(content))
		}
	}

	commitHash := objects.CreateCommit(treeHash, message, parentHash)

	// Résoudre la référence HEAD actuelle et mettre à jour
	headContent, err = GetHEAD()
	if err == nil && strings.HasPrefix(headContent, "ref: ") {
		refPath := strings.TrimPrefix(headContent, "ref: ")
		refFile := filepath.Join(".goit", refPath)
		if err := os.WriteFile(refFile, []byte(commitHash), 0644); err != nil {
			fmt.Printf("Failed to update branch reference: %v\n", err)
			return
		}
	} else {
		// Si HEAD contient directement un hash, utiliser main par défaut
		refFile := filepath.Join(".goit", "refs", "heads", "main")
		if err := os.WriteFile(refFile, []byte(commitHash), 0644); err != nil {
			fmt.Printf("Failed to update main reference: %v\n", err)
			return
		}
	}

	// Vider l'index après le commit
	os.Remove(indexPath)

	fmt.Printf("Committed: %s\n", commitHash[:8])
}

/**
 * Récupère le nom de la branche actuelle
 * Retourne "HEAD" si en état détaché
 */
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

/**
 * Met à jour la référence HEAD
 */
func SetHEAD(ref string) error {
	return os.WriteFile(".goit/HEAD", []byte(ref), 0644)
}

/**
 * Lit le contenu du fichier HEAD
 * Retourne le contenu et une erreur si problème
 */
func GetHEAD() (string, error) {
	data, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD: %v", err)
	}
	return strings.TrimSpace(string(data)), nil
}

/**
 * Récupère le hash du commit actuel
 * Résout HEAD vers le hash du commit si nécessaire
 */
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
