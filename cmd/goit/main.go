package main

import (
	"fmt"
	"os"

	"projet-go-git/internal/index"
	"projet-go-git/internal/repository"
)

func printHelp() {
	fmt.Print(`
Usage: goit <command> [options]

Commands:
	init                  Initialize a new goit repository (.goit/)
	add <file>            Add a file to the staging area
	commit -m <message>   Commit the staged changes with a message
	help                  Show this help message

Examples:
	goit init
	goit add fichier.txt
	goit commit -m "Initial commit"
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
			fmt.Println("Error: missing file to add.")
			printHelp()
			return
		}
		index.Add(os.Args[2])
	case "commit":
		if len(os.Args) < 4 || os.Args[2] != "-m" {
			fmt.Println("Error: missing or incorrect commit message.")
			printHelp()
			return
		}
		message := os.Args[3]
		repository.Commit(message)
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
	}
}
