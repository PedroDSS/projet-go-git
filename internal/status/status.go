package status

import (
	"fmt"
	"os"
	"strings"
)

func ShowStatus() {
    fmt.Println("Changes to be committed:")
    indexContent, err := os.ReadFile(".goit/index")
    if err != nil {
        fmt.Println("  (nothing added to commit)")
        return
    }

    entries := strings.Split(string(indexContent), "\n")
    for _, entry := range entries {
        if entry != "" {
            fmt.Println("  ", entry)
        }
    }
}