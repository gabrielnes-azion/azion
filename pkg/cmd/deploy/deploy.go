package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	msg "github.com/aziontech/azion-cli/messages/deploy"
	"github.com/aziontech/azion-cli/pkg/cmd/build"
	"github.com/aziontech/azion-cli/pkg/cmdutil"
	"github.com/aziontech/azion-cli/pkg/contracts"
	"github.com/aziontech/azion-cli/pkg/iostreams"
	"github.com/aziontech/azion-cli/pkg/logger"
	manifestInt "github.com/aziontech/azion-cli/pkg/manifest"
	"github.com/aziontech/azion-cli/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type DeployCmd struct {
	Io                    *iostreams.IOStreams
	GetWorkDir            func() (string, error)
	FileReader            func(path string) ([]byte, error)
	WriteFile             func(filename string, data []byte, perm fs.FileMode) error
	GetAzionJsonContent   func() (*contracts.AzionApplicationOptions, error)
	WriteAzionJsonContent func(conf *contracts.AzionApplicationOptions) error
	EnvLoader             func(path string) ([]string, error)
	BuildCmd              func(f *cmdutil.Factory) *build.BuildCmd
	Open                  func(name string) (*os.File, error)
	FilepathWalk          func(root string, fn filepath.WalkFunc) error
	F                     *cmdutil.Factory
	Unmarshal             func(data []byte, v interface{}) error
	Interpreter           func() *manifestInt.ManifestInterpreter
}

var (
	Path string
	Auto bool
)

func NewDeployCmd(f *cmdutil.Factory) *DeployCmd {
	return &DeployCmd{
		Io:                    f.IOStreams,
		GetWorkDir:            utils.GetWorkingDir,
		FileReader:            os.ReadFile,
		WriteFile:             os.WriteFile,
		EnvLoader:             utils.LoadEnvVarsFromFile,
		BuildCmd:              build.NewBuildCmd,
		GetAzionJsonContent:   utils.GetAzionJsonContent,
		WriteAzionJsonContent: utils.WriteAzionJsonContent,
		Open:                  os.Open,
		FilepathWalk:          filepath.Walk,
		Unmarshal:             json.Unmarshal,
		F:                     f,
		Interpreter:           manifestInt.NewManifestInterpreter,
	}
}

func NewCobraCmd(deploy *DeployCmd) *cobra.Command {
	deployCmd := &cobra.Command{
		Use:           msg.DeployUsage,
		Short:         msg.DeployShortDescription,
		Long:          msg.DeployLongDescription,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: heredoc.Doc(`
        $ azion deploy --help
        $ azion deploy --path dist/storage
        $ azion deploy --auto
        `),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploy.Run(deploy.F)
		},
	}
	deployCmd.Flags().BoolP("help", "h", false, msg.DeployFlagHelp)
	deployCmd.Flags().StringVar(&Path, "path", "", msg.EdgeApplicationDeployPathFlag)
	deployCmd.Flags().BoolVar(&Auto, "auto", false, msg.DeployFlagAuto)
	return deployCmd
}

func NewCmd(f *cmdutil.Factory) *cobra.Command {
	return NewCobraCmd(NewDeployCmd(f))
}

func (cmd *DeployCmd) Run(f *cmdutil.Factory) error {
	logger.Debug("Running deploy command")
	ctx := context.Background()

	// buildCmd := cmd.BuildCmd(f)
	// err := buildCmd.Run(&contracts.BuildInfo{})
	// if err != nil {
	// 	logger.Debug("Error while running build command called by deploy command", zap.Error(err))
	// 	return err
	// }

	conf, err := cmd.GetAzionJsonContent()
	if err != nil {
		logger.Debug("Failed to get Azion JSON content", zap.Error(err))
		return err
	}

	// manifest, err := readManifest(cmd)
	// if err != nil {
	// 	logger.Debug("Error while reading manifest", zap.Error(err))
	// 	return err
	// }

	clients := NewClients(f)
	interpreter := cmd.Interpreter()

	pathManifest, err := interpreter.ManifestPath()
	if err != nil {
		return err
	}

	err = cmd.doApplication(clients.EdgeApplication, context.Background(), conf)
	if err != nil {
		return err
	}

	err = cmd.doOriginSingle(clients.EdgeApplication, clients.Origin, ctx, conf)
	if err != nil {
		return err
	}

	conf.Function.File = ".edge/worker.js"
	err = cmd.doFunction(clients, ctx, conf)
	if err != nil {
		return err
	}

	manifestStructure, err := interpreter.ReadManifest(pathManifest, f)
	if err != nil {
		return err
	}

	name := ""

	if len(manifestStructure.Origins) > 0 && manifestStructure.Origins[0].Name != "" {
		name = ""
	}

	err = cmd.doBucket(clients.Bucket, ctx, conf, name)
	if err != nil {
		return err
	}

	if len(conf.RulesEngine.Rules) == 0 {
		err = cmd.doRulesDeploy(ctx, conf, clients.EdgeApplication)
		if err != nil {
			return err
		}
	}

	// skip upload when type = javascript, typescript (storage folder does not exist in these cases)
	if conf.Template != "javascript" && conf.Template != "typescript" {
		err = cmd.uploadFiles(f, conf)
		if err != nil {
			return err
		}
	}

	err = interpreter.CreateResources(conf, manifestStructure, f)
	if err != nil {
		return err
	}

	domainName, err := cmd.doDomain(clients.Domain, ctx, conf)
	if err != nil {
		return err
	}

	logger.FInfo(cmd.F.IOStreams.Out, msg.DeploySuccessful)
	logger.FInfo(cmd.F.IOStreams.Out, fmt.Sprintf(msg.DeployOutputDomainSuccess, utils.Concat("https://", domainName)))
	logger.FInfo(cmd.F.IOStreams.Out, msg.DeployPropagation)
	return nil

}
