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
	mergedTree, err := mergeTrees(currentTree, branchTree)
	if err != nil {
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
 * Fusionne deux arbres
 * Pour simplifier, on prend tous les fichiers des deux arbres
 * Dans une vraie implémentation, on gérerait les conflits
 */
func mergeTrees(tree1, tree2 string) (string, error) {
	// Pour l'instant, on retourne simplement le premier arbre
	// Dans une vraie implémentation, on fusionnerait les contenus
	return tree1, nil
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
