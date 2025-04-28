package objects

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"
)

func CreateTree() string {
    indexContent, err := os.ReadFile(".mygit/index")
    if err != nil {
        panic(err)
    }

    treeContent := "tree\n" + string(indexContent)
    hash := sha1.Sum([]byte(treeContent))
    hashStr := fmt.Sprintf("%x", hash[:])

    os.WriteFile(".mygit/objects/"+hashStr, []byte(treeContent), 0644)
    return hashStr
}

func CreateCommit(treeHash string, message string) string {
    now := time.Now().UTC().Format(time.RFC3339)
    content := fmt.Sprintf("commit\n tree %s\n date %s\n\n %s", treeHash, now, message)
    hash := sha1.Sum([]byte(content))
    hashStr := fmt.Sprintf("%x", hash[:])

    os.WriteFile(".mygit/objects/"+hashStr, []byte(content), 0644)
    return hashStr
}
