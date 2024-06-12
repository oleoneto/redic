package cli

import (
	"context"
	"time"

	"github.com/oleoneto/redic/pkg/query"
	"github.com/spf13/cobra"
)

var DefineCmd = &cobra.Command{
	Use:     "define",
	Aliases: []string{"d"},
	Args:    cobra.ExactArgs(1),
	Short:   "Get the definition of a word.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		state.BeforeHook(cmd, args)
		state.ConnectDatabase(cmd, args)
	},
	PersistentPostRun: state.AfterHook,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
		defer cancel()

		q := query.NewQuery(state.Database)

		definitions, err := q.Define(ctx, args[0], "")
		if err != nil {
			panic(err)
		}

		state.Writer.Print(definitions)
	},
}

func init() {
	DefineCmd.Flags().StringVarP(&state.Flags.DatabaseName, "database-name", "n", state.Flags.DatabaseName, "database name")

	switch state.Flags.Engine.String() {
	case "postgresql":
	default:
		DefineCmd.MarkFlagRequired("database-name")
	}
}
