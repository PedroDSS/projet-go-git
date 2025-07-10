package index

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type IndexEntry struct {
	Hash     string
	Filename string
}

/**
 * Vérifie si un merge est en cours
 */
func isMergeInProgress() bool {
	_, err := os.Stat(filepath.Join(".goit", "MERGE_HEAD"))
	return err == nil
}

/**
 * Ajoute un fichier à l'index seulement s'il a changé ou n'est pas suivi
 * Retourne true si le fichier a été ajouté, false sinon
 */
func addSingleFile(filename string, indexEntries map[string]string) (bool, error) {
	if strings.HasPrefix(filename, ".goit") {
		return false, nil
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return false, fmt.Errorf("error accessing file %s: %v", filename, err)
	}

	if fileInfo.IsDir() {
		return false, nil
	}

	// Calculer le hash actuel du fichier
	hash, content, err := processFile(filename)
	if err != nil {
		return false, err
	}

	// Cela permet de résoudre les conflits même si le fichier semble inchangé
	if isMergeInProgress() {
		// Forcer l'ajout pendant un merge
		objectPath := filepath.Join(".goit", "objects", hash)
		if err := os.WriteFile(objectPath, content, 0644); err != nil {
			return false, fmt.Errorf("error storing object %s: %v", objectPath, err)
		}
		indexEntries[filename] = hash
		return true, nil
	}

	// Vérifier si le fichier a changé par rapport à l'index
	currentHash, existsInIndex := indexEntries[filename]
	if existsInIndex && currentHash == hash {
		// Le fichier n'a pas changé par rapport à l'index, ne pas l'ajouter
		return false, nil
	}

	// Vérifier si le fichier a changé par rapport au dernier commit
	lastCommitHash := getLastCommitHash(filename)
	if lastCommitHash != "" && lastCommitHash == hash {
		// Le fichier n'a pas changé par rapport au dernier commit, ne pas l'ajouter
		return false, nil
	}

	// Le fichier a changé ou n'existe pas dans l'index/commit, l'ajouter
	objectPath := filepath.Join(".goit", "objects", hash)
	if err := os.WriteFile(objectPath, content, 0644); err != nil {
		return false, fmt.Errorf("error storing object %s: %v", objectPath, err)
	}

	indexEntries[filename] = hash
	return true, nil
}

// ... reste du code inchangé ...

/**
 * Traite un fichier pour calculer son hash et récupérer son contenu
 */
func processFile(filename string) (string, []byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", nil, fmt.Errorf("error opening file %s: %v", filename, err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", nil, fmt.Errorf("error calculating hash for %s: %v", filename, err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	if _, err := f.Seek(0, 0); err != nil {
		return "", nil, fmt.Errorf("error seeking in file %s: %v", filename, err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return "", nil, fmt.Errorf("error reading content of %s: %v", filename, err)
	}

	return hash, content, nil
}

/**
 * Charge les entrées de l'index depuis le fichier
 */
func loadIndexEntries() (map[string]string, error) {
	indexEntries := make(map[string]string)
	indexPath := filepath.Join(".goit", "index")

	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return indexEntries, nil
		}
		return nil, fmt.Errorf("error reading index: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(indexContent)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			hash, file := parts[0], parts[1]
			indexEntries[file] = hash
		}
	}

	return indexEntries, nil
}

/**
 * Écrit les entrées de l'index dans le fichier
 */
func writeIndexEntries(indexEntries map[string]string) error {
	var lines []string
	for file, hash := range indexEntries {
		lines = append(lines, fmt.Sprintf("%s %s", hash, file))
	}

	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	indexPath := filepath.Join(".goit", "index")
	return os.WriteFile(indexPath, []byte(content), 0644)
}

/**
 * Récupère le hash d'un fichier depuis le dernier commit
 */
func getLastCommitHash(filename string) string {
	// Lire HEAD
	head, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		return ""
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
		return ""
	}

	// Lire le commit
	commitPath := filepath.Join(".goit", "objects", commitHash)
	commitData, err := os.ReadFile(commitPath)
	if err != nil {
		return ""
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
		return ""
	}

	// Lire le tree
	treePath := filepath.Join(".goit", "objects", treeHash)
	treeData, err := os.ReadFile(treePath)
	if err != nil {
		return ""
	}

	// Parser le tree pour trouver le fichier
	treeLines := strings.Split(string(treeData), "\n")
	for _, line := range treeLines {
		if line != "" && !strings.HasPrefix(line, "tree") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				hash, file := parts[0], parts[1]
				if file == filename {
					return hash
				}
			}
		}
	}

	return ""
}

/**
 * Ajoute des fichiers à l'index
 * Si filename est ".", ajoute seulement les fichiers modifiés ou non suivis
 * Sinon, ajoute le fichier spécifié
 */
func Add(filename string) {
	indexEntries, err := loadIndexEntries()
	if err != nil {
		fmt.Printf("Error loading index: %v\n", err)
		return
	}

	if filename == "." {
		addedCount := 0
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if path == ".goit" || strings.HasPrefix(path, ".goit"+string(os.PathSeparator)) {
					return filepath.SkipDir
				}
				return nil
			}

			wasAdded, err := addSingleFile(path, indexEntries)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
			} else if wasAdded {
				fmt.Printf("Added %s\n", path)
				addedCount++
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Error walking directory tree: %v\n", err)
			return
		}

		if addedCount == 0 {
			fmt.Println("No files were added")
			return
		}
	} else {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Printf("pathspec '%s' did not match any files\n", filename)
			return
		}

		wasAdded, err := addSingleFile(filename, indexEntries)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if wasAdded {
			fmt.Printf("Added %s\n", filename)
		} else {
			if isMergeInProgress() {
				fmt.Printf("Added %s (merge resolution)\n", filename)
			} else {
				fmt.Printf("No changes detected for %s\n", filename)
			}
		}
	}

	if err := writeIndexEntries(indexEntries); err != nil {
		fmt.Printf("Error writing index: %v\n", err)
	}
}

/**
 * Récupère toutes les entrées de l'index
 */
func GetIndexEntries() ([]IndexEntry, error) {
	indexEntries, err := loadIndexEntries()
	if err != nil {
		return nil, err
	}

	var entries []IndexEntry
	for filename, hash := range indexEntries {
		entries = append(entries, IndexEntry{
			Hash:     hash,
			Filename: filename,
		})
	}

	return entries, nil
}

/**
 * Supprime les doublons et synchronise avec le dernier commit
 */
func CleanupAfterMerge() {
	indexEntries, err := loadIndexEntries()
	if err != nil {
		return
	}

	// Nettoyer les doublons
	cleanedEntries := make(map[string]string)
	for filename, hash := range indexEntries {
		cleanedEntries[filename] = hash
	}

	if err := writeIndexEntries(cleanedEntries); err != nil {
		fmt.Printf("Error cleaning up index: %v\n", err)
	}
}
