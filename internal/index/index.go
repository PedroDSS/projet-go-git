package index

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func Add(filename string) {
    f, err := os.Open(filename)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer f.Close()

    h := sha1.New()
    io.Copy(h, f)
    hash := fmt.Sprintf("%x", h.Sum(nil))

    f.Seek(0, 0)
    content, _ := io.ReadAll(f)

    os.WriteFile(".goit/objects/"+hash, content, 0644)
    entry := hash + " " + filename + "\n"
    os.WriteFile(".goit/index", []byte(entry), 0644)

    fmt.Println("Added", filename)
}
