package repository

import (
	"fmt"
	"os"
	"projet-go-git/internal/objects"
	"strings"
)

func Init() {
    os.Mkdir(".goit", 0755)
    os.Mkdir(".goit/objects", 0755)
    os.Mkdir(".goit/refs", 0755)
    os.WriteFile(".goit/HEAD", []byte("ref: refs/heads/master"), 0644)
    fmt.Println("Initialized empty Goit repository")
}

func Commit(message string) {
    treeHash := objects.CreateTree()
    commitHash := objects.CreateCommit(treeHash, message)
    os.WriteFile(".goit/refs/heads/master", []byte(commitHash), 0644)
    fmt.Println("Committed:", commitHash)
}

func SetHEAD(ref string) error {
	return os.WriteFile(".goit/HEAD", []byte(ref), 0644)
}

func GetHEAD() string {
	data, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}