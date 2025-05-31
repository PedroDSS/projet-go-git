package status

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Fonction pour calculer le hash d'un fichier
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

// Fonction pour vérifier si un fichier est suivi par goit
func isTracked(filename string, indexEntries map[string]string) bool {
	_, exists := indexEntries[filename]
	return exists
}

func ShowStatus() {
	// Charger l'index
	indexEntries := make(map[string]string)
	indexContent, err := os.ReadFile(".goit/index")
	if err == nil {
		// Parsing du contenu de l'index
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
	}

	// 1. Afficher les changements à commiter (fichiers dans l'index)
	fmt.Println("Changes to be committed:")
	if len(indexEntries) == 0 {
		fmt.Println("  (nothing added to commit)")
	} else {
		for filename := range indexEntries {
			fmt.Println("  new file:", filename)
		}
	}
	fmt.Println()

	// 2. Parcourir le répertoire de travail pour trouver les fichiers modifiés/non suivis
	var modified []string
	var untracked []string

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignorer les dossiers et le dossier .goit
		if info.IsDir() {
			if path == ".goit" || strings.HasPrefix(path, ".goit/") {
				return filepath.SkipDir
			}
			return nil
		}

		// Ignorer l'exécutable goit
		if path == "goit" {
			return nil
		}

		// Vérifier si le fichier est suivi
		if isTracked(path, indexEntries) {
			// Calculer le hash actuel du fichier
			currentHash, err := hashFile(path)
			if err != nil {
				return nil // Ignorer les erreurs de lecture
			}

			// Comparer avec le hash stocké
			if currentHash != indexEntries[path] {
				modified = append(modified, path)
			}
		} else {
			untracked = append(untracked, path)
		}

		return nil
	})

	// 3. Afficher les fichiers modifiés
	if len(modified) > 0 {
		fmt.Println("Changes not staged for commit:")
		for _, file := range modified {
			fmt.Println("  modified:", file)
		}
		fmt.Println()
	}

	// 4. Afficher les fichiers non suivis
	if len(untracked) > 0 {
		fmt.Println("Untracked files:")
		for _, file := range untracked {
			fmt.Println("  ", file)
		}
		fmt.Println("\nUse 'goit add <file>' to include in what will be committed")
	}
}
