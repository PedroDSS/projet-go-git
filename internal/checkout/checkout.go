package checkout

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/objects"
	"projet-go-git/internal/repository"
	"strings"
)

/**
 * Checkout change de branche et met à jour les fichiers
 * Empêche le checkout si il y a des modifications non commitées
 */
func Checkout(branchName string) error {
	// Vérifier qu'il n'y a pas de modifications non commitées
	if hasUncommittedChanges() {
		return fmt.Errorf("error: Your local changes would be overwritten by checkout.\nPlease commit your changes before switching branches.")
	}

	branchPath := filepath.Join(".goit", "refs", "heads", branchName)
	if _, err := os.Stat(branchPath); err != nil {
		return fmt.Errorf("branch %s does not exist", branchName)
	}

	// Mettre à jour HEAD
	headRef := fmt.Sprintf("ref: refs/heads/%s", branchName)
	if err := os.WriteFile(".goit/HEAD", []byte(headRef), 0644); err != nil {
		return fmt.Errorf("failed to update HEAD: %v", err)
	}

	// Mettre à jour les fichiers selon la nouvelle branche
	if err := updateWorkingDirectory(branchName); err != nil {
		return fmt.Errorf("failed to update working directory: %v", err)
	}

	fmt.Printf("Switched to branch %s\n", branchName)
	return nil
}

/**
 * Vérifie s'il y a des modifications non commitées
 */
func hasUncommittedChanges() bool {
	// Vérifier l'index
	indexPath := filepath.Join(".goit", "index")
	if _, err := os.Stat(indexPath); err == nil {
		indexContent, err := os.ReadFile(indexPath)
		if err == nil && len(strings.TrimSpace(string(indexContent))) > 0 {
			return true
		}
	}

	// Vérifier les fichiers modifiés par rapport au dernier commit
	currentHash, err := repository.GetCurrentCommitHash()
	if err != nil || currentHash == "" {
		return false // Pas de commit, donc pas de modifications
	}

	// Récupérer les fichiers du dernier commit
	commitFiles := getCommitFiles(currentHash)

	// Vérifier si des fichiers ont changé
	for filename := range commitFiles {
		if hasFileChanged(filename) {
			return true
		}
	}

	return false
}

/**
 * Récupère les fichiers d'un commit
 */
func getCommitFiles(commitHash string) map[string]string {
	files := make(map[string]string)

	// Lire le commit
	commitPath := filepath.Join(".goit", "objects", commitHash)
	commitData, err := os.ReadFile(commitPath)
	if err != nil {
		return files
	}

	// Extraire le tree hash
	lines := strings.Split(string(commitData), "\n")
	var treeHash string
	for _, line := range lines {
		if strings.HasPrefix(line, " tree ") {
			treeHash = strings.TrimPrefix(line, " tree ")
			break
		}
	}

	if treeHash == "" {
		return files
	}

	// Lire le tree
	treePath := filepath.Join(".goit", "objects", treeHash)
	treeData, err := os.ReadFile(treePath)
	if err != nil {
		return files
	}

	// Parser le tree
	treeLines := strings.Split(string(treeData), "\n")
	for _, line := range treeLines {
		if line != "" && !strings.HasPrefix(line, "tree") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				hash, filename := parts[0], parts[1]
				files[filename] = hash
			}
		}
	}

	return files
}

/**
 * Vérifie si un fichier a changé par rapport au dernier commit
 */
func hasFileChanged(filename string) bool {
	// Lire le fichier actuel
	content, err := os.ReadFile(filename)
	if err != nil {
		return false // Fichier supprimé ou inaccessible
	}

	// Calculer le hash actuel
	hash := objects.HashContent(string(content))

	// Récupérer le hash du dernier commit
	currentHash, err := repository.GetCurrentCommitHash()
	if err != nil || currentHash == "" {
		return false
	}

	commitFiles := getCommitFiles(currentHash)
	expectedHash, exists := commitFiles[filename]

	if !exists {
		return true // Nouveau fichier
	}

	return hash != expectedHash
}

/**
 * Met à jour le répertoire de travail selon la branche
 */
func updateWorkingDirectory(branchName string) error {
	// Récupérer le hash du commit de la branche
	branchPath := filepath.Join(".goit", "refs", "heads", branchName)
	branchData, err := os.ReadFile(branchPath)
	if err != nil {
		return err
	}
	branchHash := strings.TrimSpace(string(branchData))

	// Nettoyer le répertoire de travail (supprimer tous les fichiers suivis)
	if err := cleanWorkingDirectory(); err != nil {
		return err
	}

	// Restaurer les fichiers de la branche
	if err := restoreFilesFromCommit(branchHash); err != nil {
		return err
	}

	// Vider l'index
	return os.WriteFile(filepath.Join(".goit", "index"), []byte(""), 0644)
}

/**
 * Nettoie le répertoire de travail
 */
func cleanWorkingDirectory() error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignorer les dossiers spéciaux
		if info.IsDir() {
			if path == ".goit" || strings.HasPrefix(path, ".goit/") {
				return filepath.SkipDir
			}
			return nil
		}

		// Ignorer les fichiers spéciaux
		if path == "goit" || strings.HasPrefix(path, ".git") {
			return nil
		}

		// Supprimer le fichier
		return os.Remove(path)
	})
}

/**
 * Restaure les fichiers d'un commit
 */
func restoreFilesFromCommit(commitHash string) error {
	// Récupérer les fichiers du commit
	files := getCommitFiles(commitHash)

	for filename, hash := range files {
		// Lire le contenu du fichier depuis les objets
		objectPath := filepath.Join(".goit", "objects", hash)
		content, err := os.ReadFile(objectPath)
		if err != nil {
			continue
		}

		// Créer les dossiers parents si nécessaire
		dir := filepath.Dir(filename)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				continue
			}
		}

		// Écrire le fichier
		if err := os.WriteFile(filename, content, 0644); err != nil {
			continue
		}
	}

	return nil
}
