package edge_functions_instances

var (
	//domains cmd
	EdgeFuncInstanceUsage            = "edge_functions_instances"
	EdgeFuncInstanceShortDescription = "Create Edge Functions Instances for edges on Azion's platform"
	EdgeFuncInstanceLongDescription  = "Create Edge Functions Instances for edges on Azion's platform"
	EdgeFuncInstanceFlagHelp         = "Displays more information about the edge_functions_instances command"
	EdgeFuncInstanceFlagId           = "Unique identifier of the Edge Function Instance"
	ApplicationFlagId                = "Unique identifier for an edge application used by this domain.. The '--application-id' flag is mandatory"

	//delete cmd
	EdgeFuncInstanceDeleteUsage            = "delete --application-id <application_id> --function-id <function_id> [flags]"
	EdgeFuncInstanceDeleteShortDescription = "Removes an Edge Function Instance"
	EdgeFuncInstanceDeleteLongDescription  = "Removes an Edge Function Instance from the Domains library based on its given ID"
	EdgeFuncInstanceDeleteOutputSuccess    = "Edge Function Instance %s was successfully deleted\n"
	EdgeFuncInstanceDeleteHelpFlag         = "Displays more information about the delete subcommand"

	//describe cmd
	EdgeFuncInstanceDescribeUsage            = "describe --application-id <application_id> --instance-id <instance_id> [flags]"
	EdgeFuncInstanceDescribeShortDescription = "Returns the information related to the Edge Function Instance"
	EdgeFuncInstanceDescribeLongDescription  = "Returns the information related to the Edge Function Instance, informed through the flag '--instance-id' in detail"
	EdgeFuncInstanceDescribeFlagOut          = "Exports the output of the subcommand 'describe' to the given file path <file_path/file_name.ext>"
	EdgeFuncInstanceDescribeFlagFormat       = "Changes the output format passing the json value to the flag. Example '--format json'"
	EdgeFuncInstanceDescribeHelpFlag         = "Displays more information about the describe subcommand"
	EdgeFuncInstanceFileWritten              = "File successfully written to: %s\n"
)
