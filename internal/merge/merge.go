package merge

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/objects"
	"projet-go-git/internal/repository"
	"strings"
)

func Merge(branchName string) error {
	branchPath := filepath.Join(".goit", "refs", "heads", branchName)
	if _, err := os.Stat(branchPath); err != nil {
		return fmt.Errorf("branch %s does not exist", branchName)
	}

	currentBranch, err := repository.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting current branch: %v", err)
	}

	if currentBranch == branchName {
		return fmt.Errorf("cannot merge branch into itself")
	}

	currentHash, err := repository.GetCurrentCommitHash()
	if err != nil {
		return fmt.Errorf("error getting current commit: %v", err)
	}

	branchHash, err := getBranchHash(branchName)
	if err != nil {
		return fmt.Errorf("error getting branch commit: %v", err)
	}

	commonAncestor := findCommonAncestor(currentHash, branchHash)
	if commonAncestor == "" {
		return fmt.Errorf("no common ancestor found")
	}

	if commonAncestor == branchHash {
		fmt.Printf("Already up to date with %s\n", branchName)
		return nil
	}

	if commonAncestor == currentHash {
		return fastForwardMerge(branchName, branchHash)
	}

	return createMergeCommit(branchName, branchHash, currentHash)
}

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
 */
func findCommonAncestor(commit1, commit2 string) string {
	ancestors1 := getAncestors(commit1)
	ancestors2 := getAncestors(commit2)

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
	currentTree, err := getCommitTree(currentHash)
	if err != nil {
		return fmt.Errorf("error getting current tree: %v", err)
	}

	branchTree, err := getCommitTree(branchHash)
	if err != nil {
		return fmt.Errorf("error getting branch tree: %v", err)
	}

	mergedTree, err := mergeTrees(currentTree, branchTree, branchName)
	if err != nil {
		if err.Error() == "merge conflicts detected" {
			fmt.Println("Automatic merge failed\n Fix conflicts and then commit the result")
			repoRoot := findRepoRoot()
			os.WriteFile(filepath.Join(repoRoot, ".goit", "MERGE_HEAD"), []byte(branchHash), 0644)
			return nil
		}
		return fmt.Errorf("error merging trees: %v", err)
	}

	message := getMergeMessage(branchName)
	commitHash := objects.CreateCommit(mergedTree, message, currentHash)

	if err := addSecondParent(commitHash, branchHash); err != nil {
		return fmt.Errorf("error adding second parent: %v", err)
	}

	currentBranch, err := repository.GetCurrentBranch()
	if err != nil {
		return err
	}

	branchRefPath := filepath.Join(".goit", "refs", "heads", currentBranch)
	if err := os.WriteFile(branchRefPath, []byte(commitHash), 0644); err != nil {
		return fmt.Errorf("failed to update branch reference: %v", err)
	}

	syncIndexWithCommit(commitHash)

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
	files1 := getTreeFiles(tree1)
	files2 := getTreeFiles(tree2)

	mergedFiles := make(map[string]string)
	hasConflicts := false

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
			mergedFiles[filename] = hash2
		} else if !exists2 {
			mergedFiles[filename] = hash1
		} else if hash1 == hash2 {
			mergedFiles[filename] = hash1
		} else {
			mergedHash, conflict := mergeFile(filename, hash1, hash2, branchName)
			if conflict {
				hasConflicts = true
				fmt.Printf("\033[33mCONFLICT (content): Merge conflict in \033[1m%s\033[0m\033[33m\033[0m\n", filename)
			}
			mergedFiles[filename] = mergedHash
		}
	}

	treeContent := "tree\n"
	for filename, hash := range mergedFiles {
		treeContent += fmt.Sprintf("%s %s\n", hash, filename)
	}

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
	content1 := getFileContent(hash1)
	content2 := getFileContent(hash2)

	if content1 == content2 {
		return hash1, false
	}

	currentBranch, _ := repository.GetCurrentBranch()

	mergedContent := fmt.Sprintf("************** %s\n%s\n=========\n%s\n************** %s\n",
		currentBranch, content1, content2, branchName)

	mergedHash := objects.HashContent(mergedContent)
	objectPath := filepath.Join(".goit", "objects", mergedHash)
	os.WriteFile(objectPath, []byte(mergedContent), 0644)

	os.WriteFile(filename, []byte(mergedContent), 0644)

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

	lines := strings.Split(string(data), "\n")
	var newLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Insérer le parent avant la ligne vide
			newLines = append(newLines, fmt.Sprintf("parent %s", parentHash))
			newLines = append(newLines, line)
		} else {
			newLines = append(newLines, line)
		}
	}

	newData := strings.Join(newLines, "\n")
	return os.WriteFile(path, []byte(newData), 0644)
}

/**
 * Vérifie si un merge est en cours
 */
func isMergeInProgress() bool {
	repoRoot := findRepoRoot()
	_, err := os.Stat(filepath.Join(repoRoot, ".goit", "MERGE_HEAD"))
	return err == nil
}

// Trouve la racine du repo (là où il y a .goit)
func findRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, ".goit")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

/**
 * Génère le message pour un commit de merge
 */
func getMergeMessage(branchName string) string {
	return fmt.Sprintf("Merged branch '%s'", branchName)
}

/**
 * Finalise un merge après résolution des conflits
 * Crée le commit de merge final
 */
func Resolve() error {
	if !isMergeInProgress() {
		return fmt.Errorf("no merge conflicts to resolve")
	}

	repoRoot := findRepoRoot()

	indexPath := filepath.Join(repoRoot, ".goit", "index")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		os.WriteFile(indexPath, []byte(""), 0644)
	}

	// Utiliser seulement les fichiers déjà dans l'index
	treeHash := objects.CreateTree()

	currentHash, err := repository.GetCurrentCommitHash()
	if err != nil {
		return fmt.Errorf("error getting current commit: %v", err)
	}

	// Récupérer le hash de la branche mergée
	mergeHeadPath := filepath.Join(repoRoot, ".goit", "MERGE_HEAD")
	mergeHeadContent, err := os.ReadFile(mergeHeadPath)
	if err != nil {
		return fmt.Errorf("error reading MERGE_HEAD: %v", err)
	}
	branchHash := strings.TrimSpace(string(mergeHeadContent))

	// Trouver le nom de la branche
	branchName := "unknown"
	branchesDir := filepath.Join(repoRoot, ".goit", "refs", "heads")
	entries, err := os.ReadDir(branchesDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				branchPath := filepath.Join(branchesDir, entry.Name())
				branchContent, err := os.ReadFile(branchPath)
				if err == nil && strings.TrimSpace(string(branchContent)) == branchHash {
					branchName = entry.Name()
					break
				}
			}
		}
	}

	message := getMergeMessage(branchName)
	commitHash := objects.CreateCommit(treeHash, message, currentHash)

	if err := addSecondParent(commitHash, branchHash); err != nil {
		return fmt.Errorf("error adding second parent: %v", err)
	}

	// Mettre à jour la référence de la branche actuelle
	currentBranch, err := repository.GetCurrentBranch()
	if err != nil {
		return err
	}

	branchRefPath := filepath.Join(repoRoot, ".goit", "refs", "heads", currentBranch)
	if err := os.WriteFile(branchRefPath, []byte(commitHash), 0644); err != nil {
		return fmt.Errorf("failed to update branch reference: %v", err)
	}

	os.Remove(filepath.Join(repoRoot, ".goit", "MERGE_HEAD"))

	if err := syncIndexWithCommit(commitHash); err != nil {
		return fmt.Errorf("error syncing index: %v", err)
	}

	fmt.Printf("Merge completed: %s\n", commitHash[:8])
	return nil
}

func syncIndexWithCommit(commitHash string) error {
	repoRoot := findRepoRoot()
	indexPath := filepath.Join(repoRoot, ".goit", "index")

	treeHash, err := getCommitTree(commitHash)
	if err != nil {
		return fmt.Errorf("error getting commit tree: %v", err)
	}

	files := getTreeFiles(treeHash)

	var indexLines []string
	for filename, hash := range files {
		indexLines = append(indexLines, fmt.Sprintf("%s %s", hash, filename))
	}

	content := strings.Join(indexLines, "\n")
	if len(indexLines) > 0 {
		content += "\n"
	}

	return os.WriteFile(indexPath, []byte(content), 0644)
}
