package build

import "errors"

var (
	ErrorVulcanExecute         = errors.New("Error executing Vulcan: %s")
	EdgeApplicationsOutputErr  = errors.New("This output-ctrl option is not available. Read the readme files found in the repository https://github.com/aziontech/azioncli-template and try again")
	ErrFailedToRunBuildCommand = errors.New("Failed to run the build step command. Verify if the command is correct and check the output above for more details. Try the 'azioncli edge_applications build' command again or contact Azion's support")
	ErrorUnmarshalConfigFile   = errors.New("Failed to unmarshal the config.json file. Verify if the file format is JSON or fix its content according to the JSON format specification at https://www.json.org/json-en.html")
	ErrorOpeningConfigFile     = errors.New("Failed to open the config.json file. The file doesn't exist, is corrupted, or has an invalid JSON format. Verify if the file format is JSON or fix its content according to the JSON format specification at https://www.json.org/json-en.html")
	ErrorOpeningAzionFile      = errors.New("Failed to open the azion.json file. The file doesn't exist, is corrupted, or has an invalid JSON format. Verify if the file format is JSON or fix its content according to the JSON format specification at https://www.json.org/json-en.html")
	ErrorEnvFileVulcan         = errors.New("Failed to read .env file generated by Vulcan during build process")
)
