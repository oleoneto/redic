package cli

import (
	"github.com/oleoneto/redic/cmd/cli/core"
	"github.com/spf13/cobra"
)

func Execute() error {
	setupGlobalFlags()

	return RootCmd.Execute()
}

var state = core.NewCommandState()

var RootCmd = &cobra.Command{
	Use:               "redic",
	Short:             "ReDic, for when you know the words but can't quite find THE word.",
	PersistentPreRun:  state.BeforeHook,
	PersistentPostRun: state.AfterHook,
	Run:               func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func init() {
	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(SearchCmd)
	RootCmd.AddCommand(DefineCmd)
	RootCmd.AddCommand(BuildCmd)
}

func setupGlobalFlags() {
	RootCmd.PersistentFlags().VarP(state.Flags.OutputFormat, "output", "o", "output format")
	RootCmd.PersistentFlags().StringVarP(&state.Flags.OutputTemplate, "output-template", "y", state.Flags.OutputTemplate, "template (used when output format is 'gotemplate')")

	// Migrator configuration
	RootCmd.PersistentFlags().VarP(state.Flags.Engine, "adapter", "a", "database adapter")
}
