package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jedib0t/go-prompter/prompt"
)

func main() {
	p, err := prompt.New()
	if err != nil {
		log.Fatal(err)
	}

	p.SetPrefix("postgres> ")
	p.SetAutoCompleter(prompt.AutoCompleteSQLKeywords())
	var input string

	for input != "quit" {
		input, err := p.Prompt(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(input)
	}

}
