package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func isEnvMainTf(path string) bool {
	re := regexp.MustCompile(`environments/[^/]+/main\.tf$`)
	return re.MatchString(path)
}

func checkParentDirReference(file string) {
	isEnvironmentMain := isEnvMainTf(file)

	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", file)
		return
	}

	re := regexp.MustCompile(`\.\.`)
	matches := re.FindAllStringIndex(string(content), -1)

	if len(matches) > 0 {
		if isEnvironmentMain && strings.Contains(string(content), "../../modules") {
			fmt.Printf(
				"Permitted parent directory reference found in environment-specific main.tf: %s\n",
				file,
			)
		} else {
			fmt.Printf("Parent directory reference found in file: %s\n", file)
		}
	}
}

func checkTerraformDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == "modules" {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tf") {
			checkParentDirReference(path)
		}

		return nil
	})
}

var (
	dir string

	rootCmd = &cobra.Command{
		Use:   "terraform-checker",
		Short: "Terraform Checker checks for parent directory references in Terraform files",
		Run: func(cmd *cobra.Command, args []string) {
			err := checkTerraformDir(dir)
			if err != nil {
				fmt.Printf("Error walking the directory: %v\n", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&dir, "dir", "d", "", "path to the Terraform directory (required)")
	rootCmd.MarkFlagRequired("dir")
}

// Execute is the entry point of the CLI
func Execute() error {
	return rootCmd.Execute()
}
