package core

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/drewstinnett/gout/v2"
	"github.com/drewstinnett/gout/v2/config"
	"github.com/drewstinnett/gout/v2/formats/gotemplate"
	gJSON "github.com/drewstinnett/gout/v2/formats/json"
	gYAML "github.com/drewstinnett/gout/v2/formats/yaml"
	"github.com/jedib0t/go-pretty/v6/table"
	db "github.com/oleoneto/redic/db/sql"
	"github.com/oleoneto/redic/pkg/core"
	"github.com/oleoneto/redic/pkg/helpers"
	"github.com/oleoneto/redic/pkg/query"
	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagEnum struct {
	Allowed []string
	Default string
}

type CommandFlags struct {
	VerboseLogging bool
	OutputTemplate string
	OutputFormat   *FlagEnum
	Engine         *FlagEnum
	Extension      *FlagEnum
	Directory      string
	DatabaseName   string
	DatabaseURL    *string
	TimeExecutions bool
}

type CommandState struct {
	Writer             *gout.Gout
	Flags              CommandFlags
	ExecutionStartTime time.Time
	ExecutionExitLog   []any
	Database           *sql.DB
	WordNet            core.Parser
	QueryEngine        query.Query

	// Channels
	Signaler chan map[string]core.DictEntry
}

type TableFormattable interface{ TableWriter() table.Writer }

type TableFormatter struct{}

type SilentFormatter struct{}

func (f *TableFormatter) Format(data any) ([]byte, error) {
	tw, ok := data.(TableFormattable)
	if !ok {
		return []byte{}, nil
	}

	return []byte(tw.TableWriter().Render()), nil
}

func (f *SilentFormatter) Format(data any) ([]byte, error) { return []byte{}, nil }

func (ofe FlagEnum) String() string { return ofe.Default }

func (ofe *FlagEnum) Type() string { return "string" }

func (ofe *FlagEnum) Set(value string) error {
	isIncluded := func(opts []string, v string) bool {
		for _, opt := range opts {
			if v == opt {
				return true
			}
		}

		return false
	}

	if !isIncluded(ofe.Allowed, value) {
		return fmt.Errorf("%s is not a supported output format: %s", value, strings.Join(ofe.Allowed, ","))
	}

	ofe.Default = value
	return nil
}

var _ pflag.Value = (*FlagEnum)(nil)

func (c *CommandState) SetFormatter(cmd *cobra.Command, args []string) {
	switch cmd.Flag("output").Value.String() {
	case "table":
		c.Writer.SetFormatter(&TableFormatter{})
	case "json":
		c.Writer.SetFormatter(gJSON.Formatter{})
	case "yaml":
		c.Writer.SetFormatter(gYAML.Formatter{})
	case "gotemplate":
		c.Writer.SetFormatter(gotemplate.Formatter{
			Opts: config.FormatterOpts{"template": c.Flags.OutputTemplate},
		})
	case "silent":
		c.Writer.SetFormatter(&SilentFormatter{})
	case "plain":
		c.Writer.SetFormatter(PlainFormatter{})
	default:
		c.Writer.SetFormatter(gJSON.Formatter{})
	}
}

func (c *CommandState) ConnectDatabase(cmd *cobra.Command, args []string) {
	switch cmd.Flag("adapter").Value.String() {
	case "postgresql":
		if c.Flags.DatabaseURL == nil || *c.Flags.DatabaseURL == "" {
			log.Fatalln("database-url not set")
			return
		}

		var err error
		c.Database, err = db.UsePG(*c.Flags.DatabaseURL)
		if err != nil {
			log.Fatal(err)
			return
		}
	case "sqlite3":
		dbname := c.Flags.DatabaseName // i.e "redic.sqlite"
		db, err := db.UseSQLite(dbname)

		if c.Flags.VerboseLogging {
			loggerAdapter := zerologadapter.New(zerolog.New(os.Stdout))
			db = sqldblogger.OpenDriver(dbname, db.Driver(), loggerAdapter)
		}

		c.Database = db
		if c.Database == nil {
			log.Fatal(err)
			return
		}
	/*
		case "turso":
			dbname := c.Flags.DatabaseName // "redic.sqlite"
			url := "libsql://abacaxi-oleoneto.turso.io"
			authToken := os.Getenv("DATABASE_TOKEN")
			tdb := dblibsql.NewDatabase(dbname, url, authToken, 2*time.Second)

			db, err := tdb.ConnectEmbedded()

			if c.Flags.VerboseLogging {
				loggerAdapter := zerologadapter.New(zerolog.New(os.Stdout))
				db = sqldblogger.OpenDriver(fmt.Sprintf("%s?authToken=%s", url, authToken), db.Driver(), loggerAdapter)
			}

			c.Database = db

			if err != nil {
				log.Fatal(err)
				return
			}
	*/
	default:
		log.Fatal("database adapter not set")
		return
	}
}

func (c *CommandState) BeforeHook(cmd *cobra.Command, args []string) {
	if !c.Flags.TimeExecutions {
		return
	}

	c.ExecutionStartTime = time.Now()
	c.SetFormatter(cmd, args)
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

func NewCommandState() *CommandState {
	command := CommandState{
		Writer:   gout.New(),
		WordNet:  *core.DefaultParser(os.ReadDir, os.ReadFile),
		Signaler: make(chan map[string]core.DictEntry),
		Flags: CommandFlags{
			OutputFormat: &FlagEnum{
				Allowed: []string{"plain", "json", "yaml", "table", "gotemplate", "silent"},
				Default: "plain",
			},
			Engine: &FlagEnum{
				Allowed: []string{"postgresql", "sqlite3" /*,"turso"*/},
				Default: "sqlite3",
			},
			Extension: &FlagEnum{
				Allowed: []string{"yaml", "sql"},
				Default: "yaml",
			},
			DatabaseURL: helpers.PointerTo(os.Getenv("DATABASE_URL")),
		},
	}

	return &command
}

type PlainFormatter struct{}

func (w PlainFormatter) Format(v interface{}) ([]byte, error) {
	s, ok := v.(Stringable)
	if !ok {
		return []byte(fmt.Sprintf("%+v", v)), nil
	}

	return []byte(s.String()), nil
}

type Stringable interface{ String() string }
