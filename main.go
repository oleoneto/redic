package main

import (
	"embed"

	_ "github.com/joho/godotenv/autoload"
	"github.com/oleoneto/redic/cmd/cli"
)

//go:embed data
var data embed.FS

var BuildHash string = "unset"

func main() {
	cli.Execute(data, BuildHash)
}
