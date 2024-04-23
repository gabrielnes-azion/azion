package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	msg "github.com/aziontech/azion-cli/messages/deploy"
	apiEdgeApplications "github.com/aziontech/azion-cli/pkg/api/edge_applications"
	"github.com/aziontech/azion-cli/pkg/cmd/build"
	"github.com/aziontech/azion-cli/pkg/cmdutil"
	"github.com/aziontech/azion-cli/pkg/contracts"
	"github.com/aziontech/azion-cli/pkg/iostreams"
	"github.com/aziontech/azion-cli/pkg/logger"
	manifestInt "github.com/aziontech/azion-cli/pkg/manifest"
	"github.com/aziontech/azion-cli/utils"
	sdk "github.com/aziontech/azionapi-go-sdk/edgeapplications"
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

	buildCmd := cmd.BuildCmd(f)
	err := buildCmd.Run(&contracts.BuildInfo{})
	if err != nil {
		logger.Debug("Error while running build command called by deploy command", zap.Error(err))
		return err
	}

	conf, err := cmd.GetAzionJsonContent()
	if err != nil {
		logger.Debug("Failed to get Azion JSON content", zap.Error(err))
		return err
	}

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

	singleOriginId, err := cmd.doOriginSingle(clients.Origin, ctx, conf)
	if err != nil {
		return err
	}

	err = cmd.doBucket(clients.Bucket, ctx, conf)
	if err != nil {
		return err
	}

	conf.Function.File = ".edge/worker.js"
	err = cmd.doFunction(clients, ctx, conf)
	if err != nil {
		return err
	}

	if !conf.NotFirstRun {
		ruleDefaultID, err := clients.EdgeApplication.GetRulesDefault(ctx, conf.Application.ID, "request")
		if err != nil {
			logger.Debug("Error while getting default rules engine", zap.Error(err))
			return err
		}

		if strings.ToLower(conf.Preset) == "javascript" || strings.ToLower(conf.Preset) == "typescript" {
			reqRules := apiEdgeApplications.UpdateRulesEngineRequest{}
			reqRules.IdApplication = conf.Application.ID

			_, err := clients.EdgeApplication.UpdateRulesEnginePublish(ctx, &reqRules, conf.Function.InstanceID)
			if err != nil {
				return err
			}
		} else {
			behaviors := make([]sdk.RulesEngineBehaviorEntry, 0)

			var behString sdk.RulesEngineBehaviorString
			behString.SetName("set_origin")

			behString.SetTarget(strconv.Itoa(int(singleOriginId)))

			behaviors = append(behaviors, sdk.RulesEngineBehaviorEntry{
				RulesEngineBehaviorString: &behString,
			})

			reqUpdateRulesEngine := apiEdgeApplications.UpdateRulesEngineRequest{
				IdApplication: conf.Application.ID,
				Phase:         "request",
				Id:            ruleDefaultID,
			}

			reqUpdateRulesEngine.SetBehaviors(behaviors)

			_, err = clients.EdgeApplication.UpdateRulesEngine(ctx, &reqUpdateRulesEngine)
			if err != nil {
				logger.Debug("Error while updating default rules engine", zap.Error(err))
				return err
			}
		}
	}

	manifestStructure, err := interpreter.ReadManifest(pathManifest, f)
	if err != nil {
		return err
	}

	if len(conf.RulesEngine.Rules) == 0 {
		err = cmd.doRulesDeploy(ctx, conf, clients.EdgeApplication)
		if err != nil {
			return err
		}
	}

	// Check if directory exists; if not, we skip uploading static files
	if _, err := os.Stat(PathStatic); os.IsNotExist(err) {
		logger.Debug(msg.SkipUpload)
	} else {
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
