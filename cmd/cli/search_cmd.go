package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/oleoneto/redic/app"
	"github.com/oleoneto/redic/app/domain/types"
	"github.com/spf13/cobra"
)

var SearchCmd = &cobra.Command{
	Use:     "search",
	Aliases: []string{"s"},
	Args:    cobra.ArbitraryArgs,
	Short:   "Search for words matching a definition.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		state.BeforeHook(cmd, args)
		state.ConnectDatabase(cmd, args)
	},
	PersistentPostRun: state.AfterHook,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
		defer cancel()

		fmt.Println("-o", state.Flags.OutputFormat)

		// TODO: Add arguments to function call
		words, err := app.DictionaryController.FindMatchingWords(ctx, types.GetDescribedWordsInput{})
		if err != nil {
			panic(err)
		}

		state.Writer.Print(words)
	},
}
