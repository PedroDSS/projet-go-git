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

const (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
)

/**
 * Calcule le hash SHA1 d'un fichier
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
 * Récupère le nom de la branche actuelle
 */
func getCurrentBranchName() string {
	branch, err := repository.GetCurrentBranch()
	if err != nil {
		return "main"
	}
	return branch
}

/**
 * Vérifie si un merge est en cours
 */
func isMergeInProgress() bool {
	_, err := os.Stat(filepath.Join(".goit", "MERGE_HEAD"))
	return err == nil
}

/**
 * Vérifie si un répertoire doit être ignoré
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
 * Vérifie si un fichier contient des marqueurs de conflit
 */
func hasConflictMarkers(filename string) bool {
	content, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	contentStr := string(content)
	return strings.Contains(contentStr, "**************") &&
		strings.Contains(contentStr, "=========")
}

/**
 * Vérifie si un fichier a changé par rapport à sa version indexée
 */
func hasFileChanged(filename, indexHash string) bool {
	currentHash, err := hashFile(filename)
	if err != nil {
		return true // Fichier supprimé ou inaccessible
	}

	return currentHash != indexHash
}

/**
 * Vérifie si un fichier est suivi par goit
 */
func isTracked(filename string, indexEntries map[string]string, commitEntries map[string]string) bool {
	_, existsInIndex := indexEntries[filename]
	_, existsInCommit := commitEntries[filename]
	return existsInIndex || existsInCommit
}

/**
 * Récupère les fichiers du dernier commit avec leurs hashes
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
 * Charge l'index directement depuis le fichier
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
 * Afficher les fichiers avec conflits pendant un merge
 */
func showMergeConflicts() {
	var conflictFiles []string
	var resolvedFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
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

		// Vérifier si le fichier contient des marqueurs de conflit
		if hasConflictMarkers(path) {
			conflictFiles = append(conflictFiles, path)
		} else {
			// Vérifier si le fichier a été résolu (ajouté à l'index)
			indexEntries := loadIndexDirect()
			if _, exists := indexEntries[path]; exists {
				resolvedFiles = append(resolvedFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error checking for conflicts: %v\n", err)
		return
	}

	if len(conflictFiles) > 0 {
		fmt.Println("Unmerged paths:")
		for _, file := range conflictFiles {
			fmt.Printf("  %sboth modified:   %s%s\n", colorRed, file, colorReset)
		}
		fmt.Println()
	}

	if len(resolvedFiles) > 0 {
		fmt.Println("All conflicts fixed but you are still merging:")
		for _, file := range resolvedFiles {
			fmt.Printf("  %smodified:   %s%s\n", colorGreen, file, colorReset)
		}
		fmt.Println()
		fmt.Println("Use 'goit resolve' to complete the merge")
	}
}

/**
 * Affiche les différences entre les fichiers staged et working
 */
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

/**
 * Affiche le statut du repository
 */
func ShowStatus() {
	currentBranch := getCurrentBranchName()

	// Afficher le statut de merge
	if isMergeInProgress() {
		fmt.Printf("On branch %s\n", currentBranch)
		fmt.Printf("You have unmerged paths.\n")
		fmt.Printf("  (fix conflicts and run \"goit resolve\")\n")
		fmt.Printf("  (use \"goit add <file>...\" to mark resolution)\n\n")
	} else {
		fmt.Printf("On branch %s\n\n", currentBranch)
	}

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

	commitEntries := getLastCommitFiles()

	// Améliorer la détection des fichiers à commiter
	var stagedModified []string
	var stagedNew []string

	for filename, indexHash := range indexEntries {
		commitHash, existsInCommit := commitEntries[filename]
		if !existsInCommit {
			stagedNew = append(stagedNew, filename)
		} else if commitHash != indexHash {
			stagedModified = append(stagedModified, filename)
		}
	}

	// Afficher les fichiers stagés
	if len(stagedNew) > 0 || len(stagedModified) > 0 {
		fmt.Println("Changes to be committed:")
		for _, filename := range stagedNew {
			fmt.Printf("  %snew file:   %s%s\n", colorGreen, filename, colorReset)
		}
		for _, filename := range stagedModified {
			fmt.Printf("  %smodified:   %s%s\n", colorGreen, filename, colorReset)
		}
		fmt.Println()
	}

	// Pendant un merge, afficher les fichiers avec conflits
	if isMergeInProgress() {
		showMergeConflicts()
		return
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

		relPath := strings.TrimPrefix(path, "./")

		if isTracked(relPath, indexEntries, commitEntries) {
			currentHash, err := hashFile(path)
			if err != nil {
				return nil // Ignorer les erreurs de lecture
			}

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
			fmt.Printf("  %smodified:   %s%s\n", colorRed, file, colorReset)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println("Untracked files:")
		for _, file := range untracked {
			fmt.Printf("  %s%s%s\n", colorRed, file, colorReset)
		}
		fmt.Println()
		fmt.Println("Use 'goit add <file>' to include in what will be committed")
	}

	if len(stagedNew) == 0 && len(stagedModified) == 0 && len(modified) == 0 && len(untracked) == 0 {
		fmt.Println("nothing to commit, working tree clean")
	}
}
