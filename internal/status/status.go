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

/**
 * Calcule le hash SHA1 d'un fichier
 * Utilisée pour comparer les versions de fichiers
 */
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

/**
 * Récupère les fichiers du dernier commit avec leurs hashes
 * Parse le commit et le tree pour obtenir la liste des fichiers
 */
func getLastCommitFiles() map[string]string {
	commitFiles := make(map[string]string)

	head, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		return commitFiles
	}

	headContent := strings.TrimSpace(string(head))
	var commitHash string

	if strings.HasPrefix(headContent, "ref: ") {
		refPath := strings.TrimPrefix(headContent, "ref: ")
		refFile := filepath.Join(".goit", refPath)
		if content, err := os.ReadFile(refFile); err == nil {
			commitHash = strings.TrimSpace(string(content))
		}
	} else {
		commitHash = headContent
	}

	if commitHash == "" {
		return commitFiles
	}

	commitPath := filepath.Join(".goit", "objects", commitHash)
	commitData, err := os.ReadFile(commitPath)
	if err != nil {
		return commitFiles
	}

	lines := strings.Split(string(commitData), "\n")
	var treeHash string
	for _, line := range lines {
		if strings.HasPrefix(line, " tree ") {
			treeHash = strings.TrimPrefix(line, " tree ")
			break
		}
	}

	if treeHash == "" {
		return commitFiles
	}

	treePath := filepath.Join(".goit", "objects", treeHash)
	treeData, err := os.ReadFile(treePath)
	if err != nil {
		return commitFiles
	}

	treeLines := strings.Split(string(treeData), "\n")
	for _, line := range treeLines {
		if line != "" && !strings.HasPrefix(line, "tree") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				hash, filename := parts[0], parts[1]
				commitFiles[filename] = hash
			}
		}
	}

	return commitFiles
}

/**
 * Vérifie si un fichier est suivi par goit
 * Un fichier est suivi s'il est dans l'index ou dans le dernier commit
 */
func isTracked(filename string, indexEntries map[string]string, commitEntries map[string]string) bool {
	_, existsInIndex := indexEntries[filename]
	_, existsInCommit := commitEntries[filename]
	return existsInIndex || existsInCommit
}

/**
 * Affiche le statut du repository
 * Montre les fichiers staged, modifiés et non suivis
 */
func ShowStatus() {
	fmt.Printf("On branch %s\n\n", getCurrentBranchName())

	// Charger l'index
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

	// Charger les fichiers du dernier commit
	commitEntries := getLastCommitFiles()

	// Afficher les changements à commiter (fichiers dans l'index)
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

		// Normaliser le chemin (enlever le ./ au début)
		relPath := strings.TrimPrefix(path, "./")

		// Vérifier si le fichier est suivi
		if isTracked(relPath, indexEntries, commitEntries) {
			// Calculer le hash actuel du fichier
			currentHash, err := hashFile(path)
			if err != nil {
				return nil // Ignorer les erreurs de lecture
			}

			// Comparer avec le hash stocké (priorité à l'index, puis au commit)
			expectedHash := indexEntries[relPath]
			if expectedHash == "" {
				expectedHash = commitEntries[relPath]
			}

			if expectedHash != "" && currentHash != expectedHash {
				modified = append(modified, relPath)
			}
		} else {
			untracked = append(untracked, relPath)
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

/**
 * Récupère le nom de la branche actuelle
 * Utilise repository.GetCurrentBranch() pour la cohérence
 */
func getCurrentBranchName() string {
	branch, err := repository.GetCurrentBranch()
	if err != nil {
		return "main"
	}
	return branch
}

/**
 * Charge l'index directement depuis le fichier
 * Fallback si index.GetIndexEntries() échoue
 */
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

/**
 * Vérifie si un répertoire doit être ignoré
 * Ignore .goit, .git et leurs sous-répertoires
 */
func shouldIgnoreDirectory(path string) bool {
	ignoredDirs := []string{".goit", ".git"}
	for _, ignored := range ignoredDirs {
		if path == ignored || strings.HasPrefix(path, ignored+"/") {
			return true
		}
	}
	return false
}

/**
 * Vérifie si un fichier doit être ignoré
 * Ignore les fichiers système et les exécutables goit
 */
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

/**
 * Vérifie si un fichier a changé par rapport à sa version indexée
 * Compare le hash actuel avec le hash stocké dans l'index
 */
func hasFileChanged(filename, indexHash string) bool {
	currentHash, err := hashFile(filename)
	if err != nil {
		return true // Fichier supprimé ou inaccessible
	}

	return currentHash != indexHash
}

/**
 * Affiche les différences entre les fichiers staged et working
 * Peut afficher les diffs pour un fichier spécifique ou tous les fichiers
 */
func ShowDiff(filename string) {
	if filename == "" {
		showAllDiffs()
		return
	}
	showFileDiff(filename)
}

/**
 * Affiche les différences pour tous les fichiers staged
 */
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

/**
 * Affiche les différences pour un fichier spécifique
 * Compare la version staged avec la version working
 */
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
