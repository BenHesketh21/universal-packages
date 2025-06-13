package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/BenHesketh21/universal-packages/internal/oci"
	"github.com/spf13/cobra"
)

var (
	lang    string
	version string
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [package-ref]",
	Short: "Install a language package from an OCI registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ref := args[0]
		ctx := context.Background()
		fmt.Println("🧰 Pulling", ref)
		repo, err := oci.ConnectToRegistry(ref)
		if err != nil {
			fmt.Println("Error connecting to registry:", err)
			os.Exit(1)
		}
		tarball, err := oci.PullTarball(ctx, repo, ref, "./.universal-packages")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("⚒️ Downloaded package to: %s\n", tarball)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVarP(&lang, "lang", "l", "", "Package language (npm|pip|nuget)")
	installCmd.Flags().StringVarP(&version, "version", "v", "latest", "Package version")
	installCmd.MarkFlagRequired("lang")
}
