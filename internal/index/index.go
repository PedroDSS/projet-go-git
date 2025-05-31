package index

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// addSingleFile ajoute un seul fichier à l'index
func addSingleFile(filename string, indexEntries map[string]string) error {
	// Vérifier si c'est un dossier .goit
	if strings.HasPrefix(filename, ".goit") {
		return nil // Ignorer les fichiers dans .goit
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("erreur lors de l'accès au fichier %s: %v", filename, err)
	}

	if fileInfo.IsDir() {
		return nil // Ignorer les dossiers
	}

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ouverture du fichier %s: %v", filename, err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("erreur lors du calcul du hash de %s: %v", filename, err)
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// Retourner au début du fichier
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("erreur lors du repositionnement dans le fichier %s: %v", filename, err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("erreur lors de la lecture du contenu de %s: %v", filename, err)
	}

	// Écrire le contenu du fichier dans le stockage d'objets
	if err := os.WriteFile(".goit/objects/"+hash, content, 0644); err != nil {
		return fmt.Errorf("erreur lors de l'écriture dans .goit/objects/%s: %v", hash, err)
	}

	// Ajouter à la liste des entrées d'index
	indexEntries[filename] = hash
	return nil
}

// Add ajoute un fichier ou tous les fichiers (si filename est ".") à l'index
func Add(filename string) {
	// Charger l'index existant
	indexEntries := make(map[string]string)
	indexContent, err := os.ReadFile(".goit/index")
	if err == nil {
		// Analyser le contenu de l'index
		entries := strings.Split(string(indexContent), "\n")
		for _, entry := range entries {
			if entry == "" {
				continue
			}
			parts := strings.SplitN(entry, " ", 2)
			if len(parts) == 2 {
				hash, file := parts[0], parts[1]
				indexEntries[file] = hash
			}
		}
	}

	// Si le paramètre est ".", ajouter tous les fichiers
	if filename == "." {
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if path == ".goit" || path == "goit" || strings.HasPrefix(path, ".goit/") {
					return filepath.SkipDir // Ignorer le dossier .goit
				}
				return nil
			}
			if err := addSingleFile(path, indexEntries); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Added", path)
			}
			return nil
		})
		if err != nil {
			fmt.Println("Erreur lors du parcours des fichiers:", err)
		}
	} else {
		// Sinon, ajouter un seul fichier
		if err := addSingleFile(filename, indexEntries); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Added", filename)
	}

	// Écrire le nouvel index
	var newIndexContent strings.Builder
	for file, hash := range indexEntries {
		newIndexContent.WriteString(hash + " " + file + "\n")
	}
	if err := os.WriteFile(".goit/index", []byte(newIndexContent.String()), 0644); err != nil {
		fmt.Println("Erreur lors de l'écriture de l'index:", err)
	}
}
