package dev

import (
	"io"

	"github.com/MakeNowJust/heredoc"
	msg "github.com/aziontech/azion-cli/messages/dev"
	"github.com/aziontech/azion-cli/pkg/cmd/build"
	"github.com/aziontech/azion-cli/pkg/cmdutil"
	"github.com/aziontech/azion-cli/pkg/contracts"
	"github.com/aziontech/azion-cli/pkg/iostreams"
	"github.com/aziontech/azion-cli/pkg/logger"
	"github.com/aziontech/azion-cli/pkg/output"
	"github.com/aziontech/azion-cli/utils"
	"github.com/spf13/cobra"
)

var (
	isFirewall bool
)

type DevCmd struct {
	Io                    *iostreams.IOStreams
	CommandRunnerStream   func(out io.Writer, cmd string, envvars []string) error
	CommandRunInteractive func(f *cmdutil.Factory, comm string) error
	BuildCmd              func(f *cmdutil.Factory) *build.BuildCmd
	F                     *cmdutil.Factory
}

func NewDevCmd(f *cmdutil.Factory) *DevCmd {
	return &DevCmd{
		F:        f,
		Io:       f.IOStreams,
		BuildCmd: build.NewBuildCmd,
		CommandRunInteractive: func(f *cmdutil.Factory, comm string) error {
			return utils.CommandRunInteractive(f, comm)
		},
	}
}

func NewCobraCmd(dev *DevCmd) *cobra.Command {
	devCmd := &cobra.Command{
		Use:           msg.DevUsage,
		Short:         msg.DevShortDescription,
		Long:          msg.DevLongDescription,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: heredoc.Doc(`
        $ azion dev
        $ azion dev --help
        `),
		RunE: func(cmd *cobra.Command, args []string) error {
			return dev.Run(dev.F)
		},
	}
	devCmd.Flags().BoolP("help", "h", false, msg.DevFlagHelp)
	devCmd.Flags().BoolVar(&isFirewall, "firewall", false, msg.IsFirewall)
	return devCmd
}

func NewCmd(f *cmdutil.Factory) *cobra.Command {
	return NewCobraCmd(NewDevCmd(f))
}

func (cmd *DevCmd) Run(f *cmdutil.Factory) error {
	logger.Debug("Running dev command")

	if len(cmd.F.Flags.Format) > 0 {
		outGen := output.GeneralOutput{
			Msg:   "dev command is not compatible with the format flag",
			Out:   f.IOStreams.Out,
			Flags: f.Flags,
		}
		return output.Print(&outGen)
	}

	contract := &contracts.BuildInfo{}

	if isFirewall {
		contract.IsFirewall = isFirewall
		contract.OwnWorker = "true"
	}

	err := vulcan(cmd, isFirewall)
	if err != nil {
		return err
	}

	return nil
}
