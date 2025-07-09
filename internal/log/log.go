package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
	colorPink   = "\033[35m"
	colorReset  = "\033[0m"
)

type CommitInfo struct {
	Hash    string
	Message string
	Date    string
	Refs    []string
}

/**
 * Lit le fichier HEAD et retourne le nom de la branche actuelle
 * Utilisée par formatRefsWithColors() pour colorer la branche courante
 */
func getCurrentBranch() string {
	head, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		return ""
	}

	headContent := strings.TrimSpace(string(head))
	if strings.HasPrefix(headContent, "ref: ") {
		refPath := strings.TrimPrefix(headContent, "ref: ")
		if strings.HasPrefix(refPath, "refs/heads/") {
			return strings.TrimPrefix(refPath, "refs/heads/")
		}
	}
	return ""
}

/**
 * Trouve toutes les références (branches + HEAD) pointant vers un hash donné
 * Retourne une liste des noms de branches et "HEAD" si applicable
 * Utilisée par ShowLog() et ShowLogShort() pour afficher les références
 */
func getRefsForHash(targetHash string) []string {
	var refs []string

	headsDir := ".goit/refs/heads"
	if entries, err := os.ReadDir(headsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				refPath := filepath.Join(headsDir, entry.Name())
				if content, err := os.ReadFile(refPath); err == nil {
					hash := strings.TrimSpace(string(content))
					if hash == targetHash {
						refs = append(refs, entry.Name())
					}
				}
			}
		}
	}

	if head, err := os.ReadFile(".goit/HEAD"); err == nil {
		headContent := strings.TrimSpace(string(head))
		if strings.HasPrefix(headContent, "ref: ") {
			refPath := strings.TrimPrefix(headContent, "ref: ")
			refFile := filepath.Join(".goit", refPath)
			if content, err := os.ReadFile(refFile); err == nil {
				hash := strings.TrimSpace(string(content))
				if hash == targetHash {
					refs = append(refs, "HEAD")
				}
			}
		} else if headContent == targetHash {
			refs = append(refs, "HEAD")
		}
	}

	return refs
}

/**
 * Formate les références avec des couleurs pour l'affichage
 * - HEAD et branche courante : bleu et gras
 * - Autres branches : rose
 * - withParentheses : ajoute des parenthèses autour des refs
 * Utilisée par displayDetailedCommit() et displayCompactCommit()
 */
func formatRefsWithColors(refs []string, withParentheses bool) string {
	if len(refs) == 0 {
		return ""
	}

	currentBranch := getCurrentBranch()
	var headRefs, localRefs []string

	for _, ref := range refs {
		if ref == "HEAD" {
			headRefs = append(headRefs, ref)
		} else {
			localRefs = append(localRefs, ref)
		}
	}

	var result []string

	for _, ref := range headRefs {
		result = append(result, colorBlue+colorBold+ref+colorReset)
	}

	for _, ref := range localRefs {
		if ref == currentBranch {
			result = append(result, colorBlue+colorBold+ref+colorReset)
		} else {
			result = append(result, colorPink+ref+colorReset)
		}
	}

	formatted := strings.Join(result, ", ")
	if withParentheses {
		return "(" + formatted + ")"
	}
	return formatted
}

/**
 * Résout HEAD vers le hash du commit actuel
 * Lit le fichier HEAD et suit les références si nécessaire
 * Utilisée par ShowLog() et ShowLogShort() pour commencer l'affichage
 */
func getCommitHash() string {
	head, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		return ""
	}

	headContent := strings.TrimSpace(string(head))
	if strings.HasPrefix(headContent, "ref: ") {
		refPath := strings.TrimPrefix(headContent, "ref: ")
		refFile := filepath.Join(".goit", refPath)
		refContent, err := os.ReadFile(refFile)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(refContent))
	}
	return headContent
}

/**
 * Parse les données brutes d'un commit pour extraire les informations
 * Extrait le message, la date et ignore les métadonnées (tree, parent, etc.)
 * Utilisée par ShowLog() et ShowLogShort() pour traiter les données de commit
 */
func parseCommitData(data string) CommitInfo {
	lines := strings.Split(data, "\n")
	var info CommitInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, " date ") {
			dateStr := strings.TrimPrefix(line, " date ")
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				info.Date = t.Format("02/01/2006 15:04")
			} else {
				info.Date = dateStr
			}
		} else if !strings.HasPrefix(line, "commit") &&
			!strings.HasPrefix(line, " tree ") &&
			!strings.HasPrefix(line, " date ") &&
			!strings.HasPrefix(line, "parent ") &&
			line != "" {
			info.Message = line
		}
	}

	return info
}

/**
 * Extrait le hash du commit parent depuis les données du commit
 * Retourne une chaîne vide si aucun parent n'est trouvé
 * Utilisée par ShowLog() et ShowLogShort() pour naviguer dans l'historique
 */
func extractParentHash(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "parent ") {
			return strings.TrimPrefix(line, "parent ")
		}
	}
	return ""
}

/**
 * Lit le fichier d'objet commit depuis le disque
 * Helper function pour centraliser la lecture des objets commit
 * Utilisée par ShowLog() et ShowLogShort()
 */
func readCommitObject(hash string) ([]byte, error) {
	path := filepath.Join(".goit", "objects", hash)
	return os.ReadFile(path)
}

/**
 * Affiche l'historique détaillé des commits
 * Parcourt la chaîne de commits en remontant vers les parents
 * Affiche toutes les informations : hash, date, message, références
 */
func ShowLog() {
	hash := getCommitHash()
	if hash == "" {
		fmt.Println("Unable to read HEAD")
		return
	}

	for hash != "" {
		data, err := readCommitObject(hash)
		if err != nil {
			fmt.Println("Error reading commit object:", err)
			return
		}

		info := parseCommitData(string(data))
		info.Hash = hash
		info.Refs = getRefsForHash(hash)

		displayDetailedCommit(info)
		hash = extractParentHash(string(data))
	}
}

/**
 * Affiche l'historique compact des commits
 * Même logique que ShowLog() mais avec un affichage simplifié
 * Affiche seulement le hash court, le message et les références
 */
func ShowLogShort() {
	hash := getCommitHash()
	if hash == "" {
		fmt.Println("Unable to read HEAD")
		return
	}

	for hash != "" {
		data, err := readCommitObject(hash)
		if err != nil {
			fmt.Println("Error reading commit object:", err)
			return
		}

		info := parseCommitData(string(data))
		info.Hash = hash
		info.Refs = getRefsForHash(hash)

		displayCompactCommit(info)
		hash = extractParentHash(string(data))
	}
}

/**
 * Affiche un commit en format détaillé
 * Affiche toutes les informations : hash complet, date, message, références
 * Utilise formatRefsWithColors() pour colorer les références
 */
func displayDetailedCommit(info CommitInfo) {
	refsStr := formatRefsWithColors(info.Refs, false)

	fmt.Printf("%s●%s %sCommit: %s%s%s\n", colorYellow, colorReset, colorYellow, colorBold, info.Hash, colorReset)
	fmt.Printf("%s|%s %sDate:   %s%s\n", colorYellow, colorReset, colorWhite, info.Date, colorReset)
	fmt.Printf("%s|%s %sTitle:  %s%s\n", colorYellow, colorReset, colorWhite, info.Message, colorReset)
	if refsStr != "" {
		fmt.Printf("%s|%s %sRefs:   %s\n", colorYellow, colorReset, colorWhite, refsStr)
	}
	fmt.Printf("%s|%s\n", colorYellow, colorReset)
}

/**
 * Affiche un commit en formats compact
 * Affiche seulement le hash court (6 caractères), le message et les références
 * Utilise formatRefsWithColors() avec parenthèses pour les références
 */
func displayCompactCommit(info CommitInfo) {
	refsStr := formatRefsWithColors(info.Refs, true)
	shortHash := info.Hash[:6]

	if refsStr != "" {
		fmt.Printf("%s●%s %s%s%s%s %s%s%s %s\n",
			colorYellow, colorReset, colorYellow, colorBold, shortHash, colorReset,
			colorWhite, info.Message, colorReset, refsStr)
	} else {
		fmt.Printf("%s●%s %s%s%s%s %s%s%s\n",
			colorYellow, colorReset, colorYellow, colorBold, shortHash, colorReset,
			colorWhite, info.Message, colorReset)
	}
	fmt.Printf("%s|%s\n", colorYellow, colorReset)
}
