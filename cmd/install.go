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

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [package-ref]",
	Short: "Install a package from an OCI registry",
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
		client := &oci.OrasClientImpl{}
		workingDir, err := oci.Pull(ctx, client, repo, ref, "./.universal-packages")
		if err != nil {
			log.Fatal(err)
		}

		packageName, packageVersion, err := oci.GetPackageNameVersionFromRef(ref)
		if err != nil {
			log.Fatalf("could not parse package reference %q: %v", ref, err)
			os.Exit(1)
		}
		fmt.Printf("📦 Inferred package name: %s\n", packageName)

		packageType := cmd.Flag("type").Value.String()

		handler, err := packages.GetHandler(packageType)
		if err != nil {
			log.Fatalf("unsupported type %q: %v", packageType, err)
			os.Exit(1)
		}

		filePath, err := handler.LocatePackage(workingDir, packageName, packageVersion)
		if err != nil {
			log.Fatalf("could not resolve file for %q: %v", packageName, err)
			os.Exit(1)
		}

		if err := handler.UpdatePackageRef(packageName, filePath, "."); err != nil {
			log.Fatalf("error updating package reference: %v", err)
			os.Exit(1)
		}

		fmt.Printf("⚒️ Downloaded package to: %s\n", filePath)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().String("type", "", "Package type (npm|pip|nuget)")
	if err := installCmd.MarkFlagRequired("type"); err != nil {
		log.Fatalf("could not mark 'type' flag as required: %v", err)
	}
}
