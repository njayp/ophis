package basic

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "My CLI application",
	Long:  "A longer description of my CLI application",
}

var helloCmd = &cobra.Command{
	Use:   "hello [name]",
	Short: "Say hello to someone",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "World"
		if len(args) > 0 {
			name = args[0]
		}
		greeting, _ := cmd.Flags().GetString("greeting")
		cmd.Printf("%s, %s!\n", greeting, name)
	},
}

func init() {
	helloCmd.Flags().String("greeting", "Hello", "The greeting to use")
	rootCmd.AddCommand(helloCmd)
}
