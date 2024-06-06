package main

import (
	"github.com/oleoneto/redic/cmd/cli"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cli.Execute()
}
