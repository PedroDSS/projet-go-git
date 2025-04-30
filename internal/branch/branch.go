package branch

import (
	"fmt"
	"os"
	"path/filepath"
	"projet-go-git/internal/repository"
)

func Create(name string) {
	goitPath := ".goit"
	refPath := filepath.Join(goitPath, "refs", "heads", name)

	if _, err := os.Stat(refPath); err == nil {
		fmt.Println("Branch already exists:", name)
		return
	}

	head := repository.GetHEAD()
	dir := filepath.Dir(refPath)
	os.MkdirAll(dir, os.ModePerm)
	os.WriteFile(refPath, []byte(head), 0644)
	fmt.Println("Branch created:", name)
}