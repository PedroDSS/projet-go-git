package main

import (
	"fmt"
	"os"

	"projet-go-git/internal/branch"
	"projet-go-git/internal/checkout"
	"projet-go-git/internal/index"
	"projet-go-git/internal/log"
	"projet-go-git/internal/repository"
	"projet-go-git/internal/status"
)

func printHelp() {
	fmt.Print(`
Usage: goit <command> [options]

Commands:
	init                   Initialize a new goit repository (.goit/)
	add <file>             Add a file to the staging area
	commit -m <message>    Commit the staged changes with a message
	log                    Show detailed commit history
	log --compact          Show commit history compact
	status                 Show changes in the working directory
	branch <name>          List branches or create a new branch
	checkout <name>        Switch to a branch
	diff <file>            Show differences between working directory and index
	help                   Show this help message

Examples:
	goit init
	goit add fichier.txt
	goit commit -m "Initial commit"
	goit log
	goit log --compact
	goit status
	goit branch feature-1
	goit checkout feature-1
	goit diff fichier.txt
`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	// Vérifie qu'un repository est existant pour les commandes dans la liste.
	needsRepo := []string{"add", "commit", "log", "status", "branch", "checkout", "diff"}
	cmd := os.Args[1]

	for _, needRepo := range needsRepo {
		if cmd == needRepo {
			if !repository.IsGoitRepo() {
				fmt.Println("fatal: not a goit repository (or any of the parent directories)")
				os.Exit(1)
			}
			break
		}
	}

	switch cmd {
	case "init":
		repository.Init()
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: goit add <file>")
			return
		}
		index.Add(os.Args[2])
	case "commit":
		if len(os.Args) < 4 || os.Args[2] != "-m" {
			fmt.Println("Usage: goit commit -m <message>")
			return
		}
		repository.Commit(os.Args[3])
	case "log":
		if len(os.Args) > 2 && (os.Args[2] == "--compact" || os.Args[2] == "-c") {
			log.ShowLogShort()
		} else {
			log.ShowLog()
		}
	case "status":
		status.ShowStatus()
	case "branch":
		if len(os.Args) < 3 {
			branch.List()
		} else {
			branch.Create(os.Args[2])
		}
	case "checkout":
		if len(os.Args) < 3 {
			fmt.Println("Usage: goit checkout <branch>")
			return
		}
		checkout.Switch(os.Args[2])
	case "diff":
		var filename string
		if len(os.Args) >= 3 {
			filename = os.Args[2]
		}
		status.ShowDiff(filename)
	case "help":
		printHelp()
	default:
		fmt.Println("Unknown command:", os.Args[1])
		printHelp()
	}
}
