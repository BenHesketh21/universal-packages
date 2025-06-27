package cmd

import (
	"fmt"

	"github.com/BenHesketh21/universal-packages/internal/npm"
	"github.com/BenHesketh21/universal-packages/internal/oci"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push <ref>",
	Short: "Package and push a project as an OCI artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		pkgName := oci.GetPackageNameFromRef(ref)
		fmt.Printf("📦 Inferred package name: %s\n", pkgName)

		filePath, err := npm.GetLocalPackageFile(pkgName, "./package.json")
		if err != nil {
			return fmt.Errorf("could not resolve file for %q: %w", pkgName, err)
		}

		err = oci.Push(ref, filePath)
		if err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		fmt.Println("✅ Package pushed successfully!")
		return nil
	},
}

func init() {
	pushCmd.Flags().String("lang", "", "Language type (e.g., npm, pip, nuget) [required]")
	pushCmd.MarkFlagRequired("lang")
	rootCmd.AddCommand(pushCmd)
}
