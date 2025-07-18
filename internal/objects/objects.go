package objects

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"
)

/**
 * Calcule le hash SHA1 d'un contenu
 */
func HashContent(content string) string {
	hash := sha1.Sum([]byte(content))
	return fmt.Sprintf("%x", hash[:])
}

func CreateTree() string {
	indexContent, err := os.ReadFile(".goit/index")
	if err != nil {
		panic(err)
	}

	treeContent := "tree\n" + string(indexContent)
	hash := sha1.Sum([]byte(treeContent))
	hashStr := fmt.Sprintf("%x", hash[:])

	os.WriteFile(".goit/objects/"+hashStr, []byte(treeContent), 0644)
	return hashStr
}

func CreateCommit(treeHash string, message string, parentHash string) string {
	now := time.Now().UTC().Format(time.RFC3339)
	var content string
	if parentHash != "" {
		content = fmt.Sprintf("commit\n tree %s\n parent %s\n date %s\n\n %s", treeHash, parentHash, now, message)
	} else {
		content = fmt.Sprintf("commit\n tree %s\n date %s\n\n %s", treeHash, now, message)
	}
	hash := sha1.Sum([]byte(content))
	hashStr := fmt.Sprintf("%x", hash[:])

	os.WriteFile(".goit/objects/"+hashStr, []byte(content), 0644)
	return hashStr
}
