package merge

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/objects"
	"projet-go-git/internal/repository"
	"strings"
)

/**
 * Merge fusionne une branche dans la branche actuelle
 * Gère les conflits et crée un commit de merge
 */
func Merge(branchName string) error {
	// Vérifier que la branche existe
	branchPath := filepath.Join(".goit", "refs", "heads", branchName)
	if _, err := os.Stat(branchPath); err != nil {
		return fmt.Errorf("branch %s does not exist", branchName)
	}

	// Récupérer la branche actuelle
	currentBranch, err := repository.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting current branch: %v", err)
	}

	// Vérifier qu'on ne merge pas la branche actuelle
	if currentBranch == branchName {
		return fmt.Errorf("cannot merge branch into itself")
	}

	// Récupérer les hashes des commits
	currentHash, err := repository.GetCurrentCommitHash()
	if err != nil {
		return fmt.Errorf("error getting current commit: %v", err)
	}

	branchHash, err := getBranchHash(branchName)
	if err != nil {
		return fmt.Errorf("error getting branch commit: %v", err)
	}

	// Trouver l'ancêtre commun
	commonAncestor := findCommonAncestor(currentHash, branchHash)
	if commonAncestor == "" {
		return fmt.Errorf("no common ancestor found")
	}

	// Si on est déjà à jour, pas besoin de merge
	if commonAncestor == branchHash {
		fmt.Printf("Already up to date with %s\n", branchName)
		return nil
	}

	// Si on peut faire un fast-forward
	if commonAncestor == currentHash {
		return fastForwardMerge(branchName, branchHash)
	}

	// Sinon, faire un merge avec commit
	return createMergeCommit(branchName, branchHash, currentHash)
}

/**
 * Récupère le hash du commit de tête d'une branche
 */
func getBranchHash(branchName string) (string, error) {
	branchPath := filepath.Join(".goit", "refs", "heads", branchName)
	data, err := os.ReadFile(branchPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

/**
 * Trouve l'ancêtre commun entre deux commits
 * Utilise un algorithme simple de recherche
 */
func findCommonAncestor(commit1, commit2 string) string {
	// Pour simplifier, on remonte les parents jusqu'à trouver un ancêtre commun
	// Dans une vraie implémentation, on utiliserait un algorithme plus sophistiqué

	ancestors1 := getAncestors(commit1)
	ancestors2 := getAncestors(commit2)

	// Trouver le premier ancêtre commun
	for _, ancestor := range ancestors1 {
		for _, ancestor2 := range ancestors2 {
			if ancestor == ancestor2 {
				return ancestor
			}
		}
	}

	return ""
}

/**
 * Récupère tous les ancêtres d'un commit
 */
func getAncestors(commitHash string) []string {
	var ancestors []string
	hash := commitHash

	for hash != "" {
		ancestors = append(ancestors, hash)
		hash = getParentHash(hash)
	}

	return ancestors
}

/**
 * Récupère le hash du parent d'un commit
 */
func getParentHash(commitHash string) string {
	path := filepath.Join(".goit", "objects", commitHash)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "parent ") {
			return strings.TrimPrefix(line, "parent ")
		}
	}

	return ""
}

/**
 * Effectue un fast-forward merge
 * Met à jour la branche actuelle vers la branche cible
 */
func fastForwardMerge(branchName, branchHash string) error {
	// Mettre à jour la référence de la branche actuelle
	currentBranch, err := repository.GetCurrentBranch()
	if err != nil {
		return err
	}

	branchRefPath := filepath.Join(".goit", "refs", "heads", currentBranch)
	if err := os.WriteFile(branchRefPath, []byte(branchHash), 0644); err != nil {
		return fmt.Errorf("failed to update branch reference: %v", err)
	}

	fmt.Printf("Fast-forward merge: %s -> %s\n", branchName, currentBranch)
	return nil
}

/**
 * Crée un commit de merge
 * Fusionne les arbres et crée un nouveau commit avec deux parents
 */
func createMergeCommit(branchName, branchHash, currentHash string) error {
	// Récupérer les arbres des deux commits
	currentTree, err := getCommitTree(currentHash)
	if err != nil {
		return fmt.Errorf("error getting current tree: %v", err)
	}

	branchTree, err := getCommitTree(branchHash)
	if err != nil {
		return fmt.Errorf("error getting branch tree: %v", err)
	}

	// Fusionner les arbres
	mergedTree, err := mergeTrees(currentTree, branchTree, branchName)
	if err != nil {
		if err.Error() == "merge conflicts detected" {
			fmt.Println("Automatic merge failed\n Fix conflicts and then commit the result")
			return nil
		}
		return fmt.Errorf("error merging trees: %v", err)
	}

	// Créer le commit de merge
	message := fmt.Sprintf("Merge branch '%s'", branchName)
	commitHash := objects.CreateCommit(mergedTree, message, currentHash)

	// Ajouter le second parent (la branche mergée)
	if err := addSecondParent(commitHash, branchHash); err != nil {
		return fmt.Errorf("error adding second parent: %v", err)
	}

	// Mettre à jour la branche actuelle
	currentBranch, err := repository.GetCurrentBranch()
	if err != nil {
		return err
	}

	branchRefPath := filepath.Join(".goit", "refs", "heads", currentBranch)
	if err := os.WriteFile(branchRefPath, []byte(commitHash), 0644); err != nil {
		return fmt.Errorf("failed to update branch reference: %v", err)
	}

	fmt.Printf("Merge commit created: %s\n", commitHash[:8])
	return nil
}

/**
 * Récupère le hash de l'arbre d'un commit
 */
func getCommitTree(commitHash string) (string, error) {
	path := filepath.Join(".goit", "objects", commitHash)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, " tree ") {
			return strings.TrimPrefix(line, " tree "), nil
		}
	}

	return "", fmt.Errorf("no tree found in commit")
}

/**
 * Fusionne deux arbres avec gestion des conflits
 * Compare les fichiers et détecte les conflits
 */
func mergeTrees(tree1, tree2, branchName string) (string, error) {
	// Récupérer les fichiers des deux arbres
	files1 := getTreeFiles(tree1)
	files2 := getTreeFiles(tree2)

	// Fusionner les fichiers
	mergedFiles := make(map[string]string)
	hasConflicts := false

	// Traiter tous les fichiers
	allFiles := make(map[string]bool)
	for filename := range files1 {
		allFiles[filename] = true
	}
	for filename := range files2 {
		allFiles[filename] = true
	}

	for filename := range allFiles {
		hash1, exists1 := files1[filename]
		hash2, exists2 := files2[filename]

		if !exists1 {
			// Fichier seulement dans tree2 (nouveau fichier)
			mergedFiles[filename] = hash2
		} else if !exists2 {
			// Fichier seulement dans tree1 (supprimé dans tree2)
			mergedFiles[filename] = hash1
		} else if hash1 == hash2 {
			// Même contenu, pas de conflit
			mergedFiles[filename] = hash1
		} else {
			// Conflit ! Les deux branches ont modifié le fichier différemment
			mergedHash, conflict := mergeFile(filename, hash1, hash2, branchName)
			if conflict {
				hasConflicts = true
				fmt.Printf("\033[33mCONFLICT (content): Merge conflict in \033[1m%s\033[0m\033[33m\033[0m\n", filename)
			}
			mergedFiles[filename] = mergedHash
		}
	}

	// Créer le nouvel arbre
	treeContent := "tree\n"
	for filename, hash := range mergedFiles {
		treeContent += fmt.Sprintf("%s %s\n", hash, filename)
	}

	// Sauvegarder l'arbre
	treeHash := objects.HashContent(treeContent)
	treePath := filepath.Join(".goit", "objects", treeHash)
	os.WriteFile(treePath, []byte(treeContent), 0644)

	if hasConflicts {
		return "", fmt.Errorf("merge conflicts detected")
	}

	return treeHash, nil
}

/**
 * Récupère les fichiers d'un arbre
 */
func getTreeFiles(treeHash string) map[string]string {
	files := make(map[string]string)

	treePath := filepath.Join(".goit", "objects", treeHash)
	treeData, err := os.ReadFile(treePath)
	if err != nil {
		return files
	}

	lines := strings.Split(string(treeData), "\n")
	for _, line := range lines {
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
 * Fusionne un fichier avec gestion des conflits
 * Retourne le hash du fichier fusionné et s'il y a un conflit
 */
func mergeFile(filename, hash1, hash2, branchName string) (string, bool) {
	// Lire les contenus des deux versions
	content1 := getFileContent(hash1)
	content2 := getFileContent(hash2)

	// Si les contenus sont identiques, pas de conflit
	if content1 == content2 {
		return hash1, false
	}

	// Récupérer les noms des commits
	commit1Name := getCommitMessage(hash1)
	commit2Name := getCommitMessage(hash2)

	// Récupérer la branche actuelle
	currentBranch, _ := repository.GetCurrentBranch()

	// Créer le contenu avec marqueurs de conflit améliorés
	mergedContent := fmt.Sprintf("************** %s (%s)\n%s\n=========\n%s\n************** %s (%s)\n",
		currentBranch, commit1Name, content1, content2, branchName, commit2Name)

	// Sauvegarder le fichier avec conflit
	mergedHash := objects.HashContent(mergedContent)
	objectPath := filepath.Join(".goit", "objects", mergedHash)
	os.WriteFile(objectPath, []byte(mergedContent), 0644)

	// Écrire aussi dans le working directory pour que l'utilisateur puisse le voir
	os.WriteFile(filename, []byte(mergedContent), 0644)

	// Ne rien afficher ici, le nom sera dans la ligne de conflit

	return mergedHash, true
}

/**
 * Récupère le contenu d'un fichier depuis son hash
 */
func getFileContent(hash string) string {
	objectPath := filepath.Join(".goit", "objects", hash)
	content, err := os.ReadFile(objectPath)
	if err != nil {
		return ""
	}
	return string(content)
}

/**
 * Ajoute un second parent à un commit
 */
func addSecondParent(commitHash, parentHash string) error {
	path := filepath.Join(".goit", "objects", commitHash)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Ajouter la ligne parent
	newData := string(data) + fmt.Sprintf("parent %s\n", parentHash)

	return os.WriteFile(path, []byte(newData), 0644)
}

/**
 * Récupère le message d'un commit
 */
func getCommitMessage(commitHash string) string {
	objectPath := filepath.Join(".goit", "objects", commitHash)
	content, err := os.ReadFile(objectPath)
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		// Chercher la ligne qui commence par " " (espace) après les métadonnées
		if strings.HasPrefix(line, " ") && i > 0 {
			// Vérifier que la ligne précédente n'est pas une métadonnée
			prevLine := lines[i-1]
			if !strings.HasPrefix(prevLine, " tree ") &&
				!strings.HasPrefix(prevLine, " parent ") &&
				!strings.HasPrefix(prevLine, " date ") {
				return strings.TrimSpace(line)
			}
		}
	}

	return "no message"
}

/**
 * Finalise un merge après résolution des conflits
 * Crée le commit de merge final
 */
func Resolve() error {
	// Vérifier qu'on est en cours de merge (il y a des fichiers avec conflits)
	if !hasMergeConflicts() {
		return fmt.Errorf("no merge conflicts to resolve")
	}

	// Créer un tree avec les fichiers actuels
	treeHash := objects.CreateTree()

	// Récupérer le hash du commit actuel
	currentHash, err := repository.GetCurrentCommitHash()
	if err != nil {
		return fmt.Errorf("error getting current commit: %v", err)
	}

	// Créer le commit de merge
	message := "Merge conflict resolved"
	commitHash := objects.CreateCommit(treeHash, message, currentHash)

	// Mettre à jour la branche actuelle
	currentBranch, err := repository.GetCurrentBranch()
	if err != nil {
		return err
	}

	branchRefPath := filepath.Join(".goit", "refs", "heads", currentBranch)
	if err := os.WriteFile(branchRefPath, []byte(commitHash), 0644); err != nil {
		return fmt.Errorf("failed to update branch reference: %v", err)
	}

	fmt.Printf("Merge completed: %s\n", commitHash[:8])
	return nil
}

/**
 * Vérifie s'il y a des fichiers avec des marqueurs de conflit
 */
func hasMergeConflicts() bool {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

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

		// Vérifier si le fichier contient des marqueurs de conflit
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), "**************") {
			return fmt.Errorf("found conflict") // Pour arrêter la recherche
		}

		return nil
	}) != nil
}
