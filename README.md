# Projet Goit

Goit est une implémentation simplifiée de Git écrite en Go.  
Elle permet d'initialiser un dépôt local, d'ajouter des fichiers et de créer des commits, à la manière de Git.

---

## 🚀 Fonctionnalités

- Initialiser un dépôt (`init`)
- Ajouter des fichiers au staging (`add <file>`)
- Commiter les fichiers (`commit -m "<message>"`)

---

## 🛠️ Installation

### Prérequis
- Go 1.20+ installé ([Installer Go](https://golang.org/doc/install))

### Cloner le projet
```bash
git clone https://github.com/PedroDSS/projet-go-git/tree/main
cd projet-go-git
```

### Build le package (Optionnel)
```bash
go build -o goit ./cmd/goit
```

### Initialiser le repository
```bash
# Si le package a été build
./goit init
# Sinon
go run ./cmd/goit init
```

### Voir les commandes disponibles
```bash
# Si le package a été build
./goit help
# Sinon
go run ./cmd/goit help
```