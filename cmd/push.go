package cmd

import (
	"context"
	"fmt"

	"github.com/BenHesketh21/universal-packages/internal/oci"
	"github.com/BenHesketh21/universal-packages/internal/packages"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push <ref>",
	Short: "Package and push a project as an OCI artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		ctx := context.Background()
		packageName, packageVersion, err := oci.GetPackageNameVersionFromRef(ref)
		if err != nil {
			return fmt.Errorf("could not parse package reference %q: %w", ref, err)
		}
		fmt.Printf("📦 Inferred package name: %s\n", packageName)

		packageType := cmd.Flag("type").Value.String()

		handler, err := packages.GetHandler(packageType)
		if err != nil {
			return fmt.Errorf("unsupported language %q: %w", packageType, err)
		}

		filePath, err := handler.LocatePackage(".", packageName, packageVersion)
		if err != nil {
			return fmt.Errorf("could not resolve file for %q: %w", packageName, err)
		}
		client := &oci.OrasClientImpl{}
		err = oci.Push(ctx, client, ref, filePath)
		if err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		fmt.Println("✅ Package pushed successfully!")
		return nil
	},
}

func init() {
	pushCmd.Flags().String("type", "", "Package type (e.g., npm, pip, nuget) [required]")
	if err := pushCmd.MarkFlagRequired("type"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(pushCmd)
}
