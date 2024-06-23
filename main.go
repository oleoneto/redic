package main

import (
	"embed"

	_ "github.com/joho/godotenv/autoload"
	"github.com/oleoneto/redic/cmd/cli"
)

// var version = "unset"
// fmt.Println(version)
// go run -ldflags "-X main.VersionString=`git rev-parse HEAD`" main.go

//go:embed data
var data embed.FS

func main() {
	cli.Execute(data)
}
