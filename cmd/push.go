package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

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
		packageName := cmd.Flag("package-name").Value.String()
		packageVersion := cmd.Flag("package-version").Value.String()

		inferredPackageName := ""
		inferredPackageVersion := ""
		err := error(nil)
		if packageName == "" || packageVersion == "" {
			inferredPackageName, inferredPackageVersion, err = oci.GetPackageNameVersionFromRef(ref)
			if err != nil {
				log.Fatalf("could not parse package reference %q: %v", ref, err)
				os.Exit(1)
			}
		}

		if packageName == "" {
			packageName = inferredPackageName
			fmt.Printf("📦 Inferred package name: %s\n", packageName)
		}

		if packageVersion == "" {
			packageVersion = inferredPackageVersion
			fmt.Printf("📦 Inferred package version: %s\n", packageVersion)
		}

		packageType := cmd.Flag("type").Value.String()

		handler, err := packages.GetHandler(packageType)
		if err != nil {
			return fmt.Errorf("unsupported type %q: %w", packageType, err)
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
	pushCmd.Flags().String("package-name", "", "Name of the package to install, inferred from the package reference if not provided")
	pushCmd.Flags().String("package-version", "", "Version of the package to install, inferred from the package reference if not provided")
	rootCmd.AddCommand(pushCmd)
}
