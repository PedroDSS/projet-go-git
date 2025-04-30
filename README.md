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

### Installer Goit (Linux/MacOS)
## Compiler le programme
```bash
# Il faut se situer sur le chemin du dossier avant.
go build -o goit ./cmd/goit
```

## Déplacer l'executable dans le $PATH
```bash
sudo mv goit /usr/local/bin
```

## Vérifier le fonctionnement / Initialiser le repository
```bash
# Si le programme a été installé (executable partout)
goit help
# Si le package a été build
./goit init
# Sinon
go run ./cmd/goit init
```

## Voir les commandes disponibles
```bash
# Si le programme a été installé (executable partout)
goit help
# Si le package a été build
./goit help
# Sinon
go run ./cmd/goit help
```

### Désintaller
## Supprimer l'executable
```bash
sudo rm /usr/local/bin/goit
```