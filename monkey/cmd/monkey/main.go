// Package main provide a simple REPL for the monkey language.
package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/emb/play/monkey/repl"
)

func main() {
	log.SetFlags(0)
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
	fmt.Println("Feel free to play!")
	if err := repl.Start(os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
