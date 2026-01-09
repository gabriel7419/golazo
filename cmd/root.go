package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/0xjuanma/golazo/internal/app"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags
var Version = "dev"

var mockFlag bool
var updateFlag bool
var versionFlag bool
var debugFlag bool

var rootCmd = &cobra.Command{
	Use:   "golazo",
	Short: "Football match stats and updates in your terminal",
	Long:  `A modern terminal user interface for real-time football stats and scores, covering multiple leagues and competitions.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			version.Print(Version)
			return
		}

		if updateFlag {
			runUpdate()
			return
		}

		// Determine banner conditions
		isDevBuild := Version == "dev"
		newVersionAvailable := false
		storedLatestVersion := ""

		if !isDevBuild {
			if storedLatestVersion, err := data.LoadLatestVersion(); err == nil && storedLatestVersion != "" {
				// Check if new version is available (current app < stored latest)
				newVersionAvailable = version.IsOlder(Version, storedLatestVersion)
			}
		}

		// Check for updates in background (non-blocking)
		go func() {
			// Check immediately if current version is older than stored, OR do daily check
			shouldCheck := data.ShouldCheckVersion()
			if !shouldCheck && storedLatestVersion != "" && !isDevBuild {
				shouldCheck = version.IsOlder(Version, storedLatestVersion)
			}

			if shouldCheck {
				if fetchedVersion, err := data.CheckLatestVersion(); err == nil {
					data.SaveLatestVersion(fetchedVersion)
				}
			}
		}()

		p := tea.NewProgram(app.New(mockFlag, debugFlag, isDevBuild, newVersionAvailable), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
			os.Exit(1)
		}
	},
}

// runUpdate executes the install script to update golazo to the latest version.
func runUpdate() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", "irm https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.ps1 | iex")
	} else {
		cmd = exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.sh | bash")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}
}

// Execute runs the root command.
// Errors are written to stderr and the program exits with code 1 on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&mockFlag, "mock", false, "Use mock data for all views instead of real API data")
	rootCmd.Flags().BoolVar(&debugFlag, "debug", false, "Enable debug logging to ~/.golazo/golazo_debug.log")
	rootCmd.Flags().BoolVarP(&updateFlag, "update", "u", false, "Update golazo to the latest version")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Display version information")
}
