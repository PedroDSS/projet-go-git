package main

import (
	"fmt"
	"os"

	"projet-go-git/internal/index"
	"projet-go-git/internal/repository"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: goit <command>")	
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
        repository.Commit("First commit") // TODO: Dynamique.
    default:
        fmt.Println("Unknown command:", os.Args[1])
    }
}
