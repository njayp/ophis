package terraform

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func CreateTerraformCmd() *cobra.Command {
	terraformCmd := &cobra.Command{
		Use:   "terraform",
		Short: "Infrastructure as Code",
		Long:  "Terraform is a tool for building, changing, and versioning infrastructure safely and efficiently.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			runTerraformCommand(cmd, args)
		},
	}

	// Global flags that apply to all subcommands
	terraformCmd.PersistentFlags().StringP("chdir", "C", "", "Switch to a different working directory before executing the given subcommand")

	// Add popular commands
	terraformCmd.AddCommand(createInitCmd())
	terraformCmd.AddCommand(createPlanCmd())
	terraformCmd.AddCommand(createApplyCmd())
	terraformCmd.AddCommand(createDestroyCmd())
	terraformCmd.AddCommand(createValidateCmd())
	terraformCmd.AddCommand(createFmtCmd())
	terraformCmd.AddCommand(createOutputCmd())
	terraformCmd.AddCommand(createShowCmd())
	terraformCmd.AddCommand(createStateCmd())
	terraformCmd.AddCommand(createWorkspaceCmd())
	terraformCmd.AddCommand(createVersionCmd())

	return terraformCmd
}

// Helper function to run terraform commands
func runTerraformCommand(cmd *cobra.Command, args []string) {
	terraformCmd := exec.Command("terraform", args...)
	terraformCmd.Stdout = cmd.OutOrStdout()
	terraformCmd.Stderr = cmd.ErrOrStderr()
	terraformCmd.Stdin = os.Stdin
	terraformCmd.Env = os.Environ()

	if err := terraformCmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
}

func createInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Terraform configuration",
		Long:  "Initialize a working directory containing Terraform configuration files.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"init"}, args...))
		},
	}

	cmd.Flags().BoolP("upgrade", "u", false, "Upgrade modules and plugins")
	cmd.Flags().Bool("reconfigure", false, "Reconfigure backend ignoring saved configuration")
	cmd.Flags().Bool("migrate-state", false, "Allow automatic state migration")
	cmd.Flags().String("backend-config", "", "Backend configuration (can be used multiple times)")
	cmd.Flags().StringSlice("backend-config-file", nil, "Path to backend configuration file")
	cmd.Flags().Bool("get", true, "Download modules for this configuration")
	cmd.Flags().String("plugin-dir", "", "Directory containing plugin binaries")
	cmd.Flags().Bool("verify-plugins", true, "Verify plugin signatures")

	return cmd
}

func createPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [DIR]",
		Short: "Generate and show an execution plan",
		Long:  "Generate and show an execution plan for the current configuration. DIR can be specified as positional argument.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"plan"}, args...))
		},
	}

	cmd.Flags().Bool("destroy", false, "Create a plan to destroy all objects")
	cmd.Flags().Bool("detailed-exitcode", false, "Return detailed exit code (0=no changes, 1=error, 2=changes)")
	cmd.Flags().StringP("out", "o", "", "Save the plan to a file")
	cmd.Flags().Bool("refresh", true, "Update state prior to checking for differences")
	cmd.Flags().StringSliceP("target", "t", nil, "Limit planning to specific resources")
	cmd.Flags().StringSliceP("var", "", nil, "Set a variable (can be used multiple times)")
	cmd.Flags().StringSliceP("var-file", "", nil, "Load variable definitions from a file")
	cmd.Flags().Bool("input", true, "Ask for input for variables if not directly set")
	cmd.Flags().Bool("lock", true, "Lock the state file when locking is supported")
	cmd.Flags().Duration("lock-timeout", 0, "Duration to retry a state lock")
	cmd.Flags().Int("parallelism", 10, "Limit the number of concurrent operations")

	return cmd
}

func createApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [DIR or PLAN]",
		Short: "Build or change infrastructure",
		Long:  "Apply the changes required to reach the desired state of the configuration. DIR can be specified as positional argument, or a saved plan file.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"apply"}, args...))
		},
	}

	cmd.Flags().Bool("auto-approve", false, "Skip interactive approval of plan before applying")
	cmd.Flags().Bool("destroy", false, "Destroy Terraform-managed infrastructure")
	cmd.Flags().Bool("refresh", true, "Update state prior to checking for differences")
	cmd.Flags().StringSliceP("target", "t", nil, "Limit applying to specific resources")
	cmd.Flags().StringSliceP("var", "", nil, "Set a variable (can be used multiple times)")
	cmd.Flags().StringSliceP("var-file", "", nil, "Load variable definitions from a file")
	cmd.Flags().Bool("input", true, "Ask for input for variables if not directly set")
	cmd.Flags().Bool("lock", true, "Lock the state file when locking is supported")
	cmd.Flags().Duration("lock-timeout", 0, "Duration to retry a state lock")
	cmd.Flags().Int("parallelism", 10, "Limit the number of concurrent operations")
	cmd.Flags().String("state", "", "Path to read and save state")
	cmd.Flags().String("state-out", "", "Path to write state to that is different from input")

	return cmd
}

func createDestroyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy previously-created infrastructure",
		Long:  "Destroy the Terraform-managed infrastructure.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"destroy"}, args...))
		},
	}

	cmd.Flags().Bool("auto-approve", false, "Skip interactive approval before destroying")
	cmd.Flags().Bool("refresh", true, "Update state prior to checking for differences")
	cmd.Flags().StringSliceP("target", "t", nil, "Limit destroying to specific resources")
	cmd.Flags().StringSliceP("var", "", nil, "Set a variable (can be used multiple times)")
	cmd.Flags().StringSliceP("var-file", "", nil, "Load variable definitions from a file")
	cmd.Flags().Bool("input", true, "Ask for input for variables if not directly set")
	cmd.Flags().Bool("lock", true, "Lock the state file when locking is supported")
	cmd.Flags().Duration("lock-timeout", 0, "Duration to retry a state lock")
	cmd.Flags().Int("parallelism", 10, "Limit the number of concurrent operations")
	cmd.Flags().String("state", "", "Path to read and save state")

	return cmd
}

func createValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [DIR]",
		Short: "Validate the Terraform files",
		Long:  "Validate the configuration files in a directory. DIR can be specified as positional argument.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"validate"}, args...))
		},
	}

	cmd.Flags().Bool("json", false, "Produce output in JSON format")
	cmd.Flags().Bool("no-color", false, "Disable colorized output")

	return cmd
}

func createFmtCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt [DIR]",
		Short: "Rewrite Terraform configuration files to canonical format",
		Long:  "Rewrite Terraform configuration files to a canonical format and style. DIR can be specified as positional argument.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"fmt"}, args...))
		},
	}

	cmd.Flags().Bool("list", true, "List files whose formatting differs")
	cmd.Flags().Bool("write", true, "Write result to source file instead of STDOUT")
	cmd.Flags().Bool("diff", false, "Display diffs of formatting changes")
	cmd.Flags().Bool("check", false, "Check if the input is formatted (exit status 1 if not)")
	cmd.Flags().Bool("recursive", false, "Also process files in subdirectories")

	return cmd
}

func createOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Show output values from your root module",
		Long:  "Show output values from your root module.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"output"}, args...))
		},
	}

	cmd.Flags().Bool("json", false, "Machine readable output in JSON format")
	cmd.Flags().Bool("raw", false, "For value types that can be automatically converted to a string")
	cmd.Flags().String("state", "", "Path to the state file to read")
	cmd.Flags().Bool("no-color", false, "Disable colorized output")

	return cmd
}

func createShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the current state or a saved plan",
		Long:  "Show the current state or a saved plan.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"show"}, args...))
		},
	}

	cmd.Flags().Bool("json", false, "Machine readable output in JSON format")
	cmd.Flags().Bool("no-color", false, "Disable colorized output")

	return cmd
}

func createStateCmd() *cobra.Command {
	stateCmd := &cobra.Command{
		Use:   "state <subcommand>",
		Short: "Advanced state management",
		Long:  "Advanced state management commands.",
	}

	// state list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List resources in the state",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"state", "list"}, args...))
		},
	}
	listCmd.Flags().String("state", "", "Path to a Terraform state file")
	listCmd.Flags().String("id", "", "Filters the results by resource instance ID")

	// state show
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show a resource in the state",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"state", "show"}, args...))
		},
	}
	showCmd.Flags().String("state", "", "Path to a Terraform state file")

	// state rm
	rmCmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove instances from the state",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"state", "rm"}, args...))
		},
	}
	rmCmd.Flags().String("state", "", "Path to the state file to update")
	rmCmd.Flags().String("backup", "", "Path to backup the existing state file")
	rmCmd.Flags().Bool("dry-run", false, "Only print what would be removed")
	rmCmd.Flags().Bool("lock", true, "Lock the state file when locking is supported")
	rmCmd.Flags().Duration("lock-timeout", 0, "Duration to retry a state lock")

	// state mv
	mvCmd := &cobra.Command{
		Use:   "mv",
		Short: "Move an item in the state",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"state", "mv"}, args...))
		},
	}
	mvCmd.Flags().String("state", "", "Path to the source state file")
	mvCmd.Flags().String("state-out", "", "Path to the destination state file")
	mvCmd.Flags().String("backup", "", "Path to backup the existing state file")
	mvCmd.Flags().Bool("dry-run", false, "Only print what would be moved")
	mvCmd.Flags().Bool("lock", true, "Lock the state file when locking is supported")
	mvCmd.Flags().Duration("lock-timeout", 0, "Duration to retry a state lock")

	stateCmd.AddCommand(listCmd, showCmd, rmCmd, mvCmd)
	return stateCmd
}

func createWorkspaceCmd() *cobra.Command {
	workspaceCmd := &cobra.Command{
		Use:   "workspace <subcommand>",
		Short: "Workspace management",
		Long:  "Workspace management commands.",
	}

	// workspace list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Workspaces",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"workspace", "list"}, args...))
		},
	}

	// workspace select
	selectCmd := &cobra.Command{
		Use:   "select",
		Short: "Select a workspace",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"workspace", "select"}, args...))
		},
	}

	// workspace new
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new workspace",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"workspace", "new"}, args...))
		},
	}
	newCmd.Flags().String("state", "", "Copy an existing state file into the new workspace")

	// workspace delete
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a workspace",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"workspace", "delete"}, args...))
		},
	}
	deleteCmd.Flags().Bool("force", false, "Remove a non-empty workspace")
	deleteCmd.Flags().Bool("lock", true, "Lock the state file when locking is supported")
	deleteCmd.Flags().Duration("lock-timeout", 0, "Duration to retry a state lock")

	// workspace show
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show the name of the current workspace",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"workspace", "show"}, args...))
		},
	}

	workspaceCmd.AddCommand(listCmd, selectCmd, newCmd, deleteCmd, showCmd)
	return workspaceCmd
}

func createVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the current Terraform version",
		Long:  "Show the current Terraform version and available provider versions.",
		Run: func(cmd *cobra.Command, args []string) {
			runTerraformCommand(cmd, append([]string{"version"}, args...))
		},
	}

	cmd.Flags().Bool("json", false, "Machine readable output in JSON format")

	return cmd
}
