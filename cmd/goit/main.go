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
	log                    Show commit history
	status                 Show changes in the working directory
	branch <name>          Create a new branch
	checkout <name>        Switch to a branch
	help                   Show this help message

Examples:
	goit init
	goit add fichier.txt
	goit commit -m "Initial commit"
	goit log
	goit status
	goit branch feature-1
	goit checkout feature-1
`)
}


func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
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
		log.ShowLog()
	case "status":
		status.ShowStatus()
	case "branch":
		if len(os.Args) < 3 {
			fmt.Println("Usage: goit branch <name>")
			return
		}
		branch.Create(os.Args[2])
	case "checkout":
		if len(os.Args) < 3 {
			fmt.Println("Usage: goit checkout <branch>")
			return
		}
		checkout.Switch(os.Args[2])
	default:
		fmt.Println("Commande inconnue:", os.Args[1])
		printHelp()
	}
}
