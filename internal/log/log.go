package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ShowLog() {
	head, err := os.ReadFile(".goit/HEAD")
	if err != nil {
		fmt.Println("Unable to read HEAD")
		return
	}

	hash := string(head)
	for hash != "" {
		path := filepath.Join(".goit", "objects", hash)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Error reading commit object:", err)
			return
		}

		lines := string(data)
		fmt.Println("Commit:", hash)
		fmt.Println(lines)
		fmt.Println("----------------------")

		hash = extractParentHash(lines)
	}
}

func extractParentHash(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "parent ") {
			return strings.TrimPrefix(line, "parent ")
		}
	}
	return ""
}