package root

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	msg "github.com/aziontech/azion-cli/messages/root"
	buildCmd "github.com/aziontech/azion-cli/pkg/cmd/build"
	"github.com/aziontech/azion-cli/pkg/cmd/completion"
	"github.com/aziontech/azion-cli/pkg/cmd/create"
	"github.com/aziontech/azion-cli/pkg/cmd/delete"
	"github.com/aziontech/azion-cli/pkg/cmd/describe"
	"github.com/aziontech/azion-cli/pkg/cmd/list"
	"github.com/aziontech/azion-cli/pkg/cmd/login"
	"github.com/aziontech/azion-cli/pkg/cmd/logout"
	logcmd "github.com/aziontech/azion-cli/pkg/cmd/logs"
	"github.com/aziontech/azion-cli/pkg/cmd/purge"
	"github.com/aziontech/azion-cli/pkg/cmd/reset"
	"github.com/aziontech/azion-cli/pkg/cmd/sync"
	"github.com/aziontech/azion-cli/pkg/cmd/unlink"
	"github.com/aziontech/azion-cli/pkg/cmd/update"
	"github.com/aziontech/azion-cli/pkg/cmd/whoami"
	"github.com/aziontech/azion-cli/pkg/metric"
	"github.com/aziontech/azion-cli/pkg/output"
	"github.com/aziontech/azion-cli/pkg/schedule"

	deploycmd "github.com/aziontech/azion-cli/pkg/cmd/deploy"
	devcmd "github.com/aziontech/azion-cli/pkg/cmd/dev"
	initcmd "github.com/aziontech/azion-cli/pkg/cmd/init"
	linkcmd "github.com/aziontech/azion-cli/pkg/cmd/link"
	"github.com/aziontech/azion-cli/pkg/cmd/version"
	"github.com/aziontech/azion-cli/pkg/cmdutil"
	"github.com/aziontech/azion-cli/pkg/constants"
	"github.com/aziontech/azion-cli/pkg/iostreams"
	"github.com/aziontech/azion-cli/pkg/logger"
	"github.com/aziontech/azion-cli/pkg/token"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RootCmd struct {
	F *cmdutil.Factory
}

func NewRootCmd(f *cmdutil.Factory) *RootCmd {
	return &RootCmd{
		F: f,
	}
}

var (
	tokenFlag      string
	configFlag     string
	commandName    string
	globalSettings *token.Settings
	startTime      time.Time
)

const PREFIX_FLAG = "--"

func NewCmd(f *cmdutil.Factory) *cobra.Command {
	return NewCobraCmd(NewRootCmd(f), f)
}

func NewCobraCmd(rootCmd *RootCmd, f *cmdutil.Factory) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:     msg.RootUsage,
		Long:    msg.RootDescription,
		Short:   color.New(color.Bold).Sprint(fmt.Sprintf(msg.RootDescription, version.BinVersion)),
		Version: version.BinVersion,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			logger.LogLevel(f.Logger)

			if strings.HasPrefix(configFlag, PREFIX_FLAG) {
				return msg.ErrorPrefix
			}

			if err := doPreCommandCheck(cmd, f, PreCmd{
				config: configFlag,
				token:  tokenFlag,
			}); err != nil {
				return err
			}

			schedule.ExecSchedules(f)
			return nil
		},
		Example: heredoc.Doc(`
		$ azion
		$ azion -t azionb43a9554776zeg05b11cb1declkbabcc9la
		$ azion --debug
		$ azion -h
		`),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().Changed("token") {
				return nil
			}
			return cmd.Help()
		},
		SilenceErrors: true, // Silence errors, so the help message won't be shown on flag error
		SilenceUsage:  true, // Silence usage on error
	}

	cobraCmd.SetIn(f.IOStreams.In)
	cobraCmd.SetOut(f.IOStreams.Out)
	cobraCmd.SetErr(f.IOStreams.Err)

	cobraCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		rootHelpFunc(cmd, args)
	})

	// Global flags
	cobraCmd.PersistentFlags().StringVarP(&tokenFlag, "token", "t", "", msg.RootTokenFlag)
	cobraCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", msg.RootConfigFlag)
	cobraCmd.PersistentFlags().BoolVarP(&f.Debug, "debug", "d", false, msg.RootLogDebug)
	cobraCmd.PersistentFlags().BoolVarP(&f.Silent, "silent", "s", false, msg.RootLogSilent)
	cobraCmd.PersistentFlags().StringVarP(&f.LogLevel, "log-level", "l", "info", msg.RootLogLevel)
	cobraCmd.PersistentFlags().BoolVarP(&f.GlobalFlagAll, "yes", "y", false, msg.RootYesFlag)
	cobraCmd.PersistentFlags().StringVar(&f.Out, "out", "", msg.RootFlagOut)
	cobraCmd.PersistentFlags().StringVar(&f.Format, "format", "", msg.RootFlagFormat)
	cobraCmd.PersistentFlags().BoolVar(&f.NoColor, "no-color", false, msg.RootFlagFormat)

	// other flags
	cobraCmd.Flags().BoolP("help", "h", false, msg.RootHelpFlag)

	// set template for -v flag
	cobraCmd.SetVersionTemplate(color.New(color.Bold).Sprint("Azion CLI " + version.BinVersion + "\n"))

	cobraCmd.AddCommand(initcmd.NewCmd(f))
	cobraCmd.AddCommand(logcmd.NewCmd(f))
	cobraCmd.AddCommand(deploycmd.NewCmd(f))
	cobraCmd.AddCommand(buildCmd.NewCmd(f))
	cobraCmd.AddCommand(devcmd.NewCmd(f))
	cobraCmd.AddCommand(linkcmd.NewCmd(f))
	cobraCmd.AddCommand(unlink.NewCmd(f))
	cobraCmd.AddCommand(completion.NewCmd(f))
	cobraCmd.AddCommand(describe.NewCmd(f))
	cobraCmd.AddCommand(login.NewCmd(f))
	cobraCmd.AddCommand(logout.NewCmd(f))
	cobraCmd.AddCommand(create.NewCmd(f))
	cobraCmd.AddCommand(list.NewCmd(f))
	cobraCmd.AddCommand(delete.NewCmd(f))
	cobraCmd.AddCommand(update.NewCmd(f))
	cobraCmd.AddCommand(version.NewCmd(f))
	cobraCmd.AddCommand(whoami.NewCmd(f))
	cobraCmd.AddCommand(purge.NewCmd(f))
	cobraCmd.AddCommand(reset.NewCmd(f))
	cobraCmd.AddCommand(sync.NewCmd(f))

	return cobraCmd
}

func Execute() {
	logger.New(zapcore.InfoLevel)

	streams := iostreams.System()
	httpClient := &http.Client{
		Timeout: 10 * time.Second, // TODO: Configure this somewhere
	}

	tok, _ := token.ReadSettings()
	viper.SetEnvPrefix("AZIONCLI")
	viper.AutomaticEnv()
	viper.SetDefault("token", tok.Token)
	viper.SetDefault("api_url", constants.ApiURL)
	viper.SetDefault("storage_url", constants.StorageApiURL)

	factory := &cmdutil.Factory{
		HttpClient: httpClient,
		IOStreams:  streams,
		Config:     viper.GetViper(),
	}

	cmd := NewCmd(factory)
	err := cmd.Execute()
	executionTime := time.Since(startTime).Seconds()

	// 1 = authorize; anything different than 1 means that the user did not authorize metrics collection, or did not answer the question yet
	if globalSettings != nil {
		if globalSettings.AuthorizeMetricsCollection == 1 {
			errMetrics := metric.TotalCommandsCount(cmd, commandName, executionTime, err)
			if errMetrics != nil {
				logger.Debug("Error while saving metrics", zap.Error(err))
			}
		}
	}

	if err != nil {
		output.Print(&output.ErrorOutput{
			GeneralOutput: output.GeneralOutput{
				Out:   streams.Out,
				Flags: factory.Flags,
			},
			Err: err,
		})
	}
}
