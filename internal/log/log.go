package log

import (
	"fmt"
	"projet-go-git/internal/objects"
	"projet-go-git/internal/repository"
	"strings"
)

func ShowLog() {
	currentHash, err := repository.GetCurrentCommitHash()
	if err != nil {
		fmt.Printf("Failed to get current commit: %v\n", err)
		return
	}

	if currentHash == "" {
		fmt.Println("No commits yet")
		return
	}

	hash := currentHash
	for hash != "" {
		content, err := objects.ReadObject(hash)
		if err != nil {
			fmt.Printf("Error reading commit %s: %v\n", hash[:8], err)
			return
		}

		fmt.Printf("commit %s\n", hash)
		printCommitInfo(content)
		fmt.Println("----------------------------------------")

		hash = extractParentHash(content)
	}
}

func printCommitInfo(content string) {
	lines := strings.Split(content, "\n")
	inMessage := false

	for _, line := range lines {
		if strings.HasPrefix(line, "date ") {
			fmt.Printf("Date: %s\n", strings.TrimPrefix(line, "date "))
		} else if line == "" && !inMessage {
			inMessage = true
			fmt.Println()
		} else if inMessage {
			fmt.Printf("    %s\n", line)
		}
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
