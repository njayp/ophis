package basic

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	helloCmd := &cobra.Command{
		Use:   "hello [name]",
		Short: "Say hello to someone",
		Long:  "Say hello to someone using a customizable greeting. The name is optional and defaults to 'World'.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := "World"
			if len(args) > 0 {
				name = args[0]
			}
			// Try to get from parent's persistent flags
			greeting, _ := cmd.Parent().PersistentFlags().GetString("greeting")
			if greeting == "" {
				greeting = "Hello"
			}
			// Debug output
			cmd.Printf("%s, %s!\n", greeting, name)
		},
	}

	rootCmd := &cobra.Command{
		Use:   "myapp",
		Short: "My CLI application",
		Long:  "A longer description of my CLI application",
	}

	rootCmd.PersistentFlags().String("greeting", "Hello", "The greeting to use")
	// Also add the greeting flag directly to the hello command for testing
	helloCmd.Flags().String("greeting", "Hello", "The greeting to use")
	rootCmd.AddCommand(helloCmd)

	return rootCmd
}
