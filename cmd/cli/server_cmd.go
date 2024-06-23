package cli

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oleoneto/redic/app"
	"github.com/oleoneto/redic/app/domain/external"
	"github.com/oleoneto/redic/cmd/api"
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use:              "server",
	Short:            "Redic API server",
	PersistentPreRun: state.ConnectDatabase,
	Run: func(cmd *cobra.Command, args []string) {
		app.New(external.DBConnectOptions{DB: state.Database})

		api.CreateAPI(fiber.Config{}).
			Listen(state.Flags.ServerAddr)
	},
}

func init() {
	ServerCmd.Flags().StringVar(&state.Flags.ServerAddr, "address", state.Flags.ServerAddr, "")
}
