package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/oleoneto/redic/cmd/cli"
)

func main() {
	cli.Execute()
}
