package cli

import (
	"context"
	"time"

	"github.com/oleoneto/redic/app"
	"github.com/oleoneto/redic/app/domain/types"
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

		// TODO: Review arguments to function call
		definitions, err := app.DictionaryController.GetDefinition(ctx, types.GetWordDefinitionsInput{Word: args[0], PartOfSpeech: "*"})
		if err != nil {
			panic(err)
		}

		state.Writer.Print(definitions)
	},
}
