package core

import (
	"fmt"
	"os"
	"time"

	"github.com/oleoneto/redic/app/domain/external"
	dbsql "github.com/oleoneto/redic/app/repositories/sql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *CommandState) ConnectDatabase(cmd *cobra.Command, args []string) {
	dbpath := viper.Get("database.path")

	db, err := dbsql.ConnectDatabase(external.DBConnectOptions{
		Adapter:  external.SQLAdapter(cmd.Flag("adapter").Value.String()),
		DSN:      *c.Flags.DatabaseURL,
		Filename: dbpath.(string),
	})
	if err != nil {
		panic(err)
	}

	c.Database = db
}

func (c *CommandState) BeforeHook(cmd *cobra.Command, args []string) {
	c.SetFormatter(cmd, args)

	if !c.Flags.TimeExecutions {
		return
	}

	c.ExecutionStartTime = time.Now()
}

func (c *CommandState) AfterHook(cmd *cobra.Command, args []string) {
	if !c.Flags.TimeExecutions {
		return
	}

	fmt.Fprintln(
		os.Stderr,
		append([]any{"Elapsed time:", time.Since(c.ExecutionStartTime)}, c.ExecutionExitLog...)...,
	)
}
