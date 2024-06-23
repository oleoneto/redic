package cli

import (
	"embed"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/oleoneto/go-toolkit/files"
	"github.com/oleoneto/redic/cmd/cli/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// wherein dictionary files are stored
var virtualFS embed.FS

var dbfile = "data/redic.sqlite"

var state = core.NewCommandState()

var RootCmd = &cobra.Command{
	Use:               "redic",
	Short:             "ReDic, for when you know the words but can't quite find THE word.",
	PersistentPreRun:  state.BeforeHook,
	PersistentPostRun: state.AfterHook,
	Run:               func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the CLI by creating its required configuration files",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		state.Flags.HomeDirectory, err = homedir.Dir()
		if err != nil {
			log.Fatalln(err)
		}

		state.Flags.ConfigDir.Create(files.FileGenerator{}, state.Flags.HomeDirectory)

		// NOTE: Copy embeded dictionary to local filesystem

		dictionary, err := virtualFS.Open(dbfile)
		if err != nil {
			log.Fatalln(err)
		}

		dstDir := fmt.Sprintf("%s/%s/%s", state.Flags.HomeDirectory, state.Flags.ConfigDir.Name, dbfile)
		dstFile, err := os.Create(dstDir)
		if err != nil {
			log.Fatalln(err)
		}

		if _, err := io.Copy(dstFile, dictionary); err != nil {
			log.Fatalln(err)
		}

		viper.Set("database.path", dstFile.Name())
		viper.WriteConfig()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(SearchCmd)
	RootCmd.AddCommand(DefineCmd)
	RootCmd.AddCommand(CreateTablesCmd)
	RootCmd.AddCommand(ReindexCmd)
	RootCmd.AddCommand(ServerCmd)
}

func initConfig() {
	if state.Flags.CLIConfig != "" {
		viper.SetConfigFile(state.Flags.CLIConfig)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalln(err)
		}

		path := fmt.Sprintf("%v/%s", home, state.Flags.ConfigDir.Name)

		viper.AddConfigPath(path)
		viper.SetConfigName("config")
	}

	// NOTE: File does not exist... create one!
	if err := viper.ReadInConfig(); err != nil {
		home, herr := homedir.Dir()
		if herr != nil {
			log.Fatalln(err)
		}

		if f := state.Flags.ConfigDir.Create(files.FileGenerator{}, home); len(f) == 0 {
			log.Fatalln("Cannot read config. Hint: You may need to run `init` to create the config file")
		}
	}
}

func Execute(vfs embed.FS) error {
	virtualFS = vfs

	// MARK: Set up global glags
	RootCmd.PersistentFlags().BoolVar(&state.Flags.VerboseLogging, "verbose", state.Flags.VerboseLogging, "enable detailed logging")
	RootCmd.PersistentFlags().VarP(state.Flags.OutputFormat, "output", "o", "output format")
	RootCmd.PersistentFlags().StringVarP(&state.Flags.OutputTemplate, "output-template", "y", state.Flags.OutputTemplate, "template (used when output format is 'gotemplate')")

	// Migrator configuration
	RootCmd.PersistentFlags().VarP(state.Flags.Engine, "adapter", "a", "database adapter")
	RootCmd.PersistentFlags().BoolVar(&state.Flags.TimeExecutions, "time", state.Flags.TimeExecutions, "time executions")

	// MARK: Run
	return RootCmd.Execute()
}
