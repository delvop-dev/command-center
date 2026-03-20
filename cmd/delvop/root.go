package delvop

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/delvop-dev/delvop/internal/config"
	"github.com/delvop-dev/delvop/internal/governance"
	"github.com/delvop-dev/delvop/internal/hooks"
	"github.com/delvop-dev/delvop/internal/notify"
	"github.com/delvop-dev/delvop/internal/security"
	"github.com/delvop-dev/delvop/internal/session"
	"github.com/delvop-dev/delvop/internal/tui"

	// Register all providers via init()
	_ "github.com/delvop-dev/delvop/internal/provider"
)

var (
	appVersion = "dev"

	rootCmd = &cobra.Command{
		Use:   "delvop",
		Short: "Engineering Command Center for Terminal Coding Agents",
		Long:  "Manage a team of AI coding agents from a single terminal dashboard.",
		RunE:  runDashboard,
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version of delvop",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("delvop", appVersion)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func SetVersion(v string) {
	appVersion = v
}

func Execute() error {
	return rootCmd.Execute()
}

func runDashboard(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config load error: %v, using defaults\n", err)
		cfg = config.Default()
	}

	hookEngine := hooks.New("/tmp/delvop.sock")
	if err := hookEngine.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: hook engine start error: %v\n", err)
	}
	defer hookEngine.Stop()

	notifier := notify.New(cfg)
	gov, _ := governance.Load()
	scanner := security.New()
	if gov != nil {
		for _, id := range gov.Security.DisabledRules {
			scanner.DisableRule(id)
		}
		scanner.SetAllowedHosts(gov.Security.CustomAllowedHosts)
	}
	mgr := session.NewManager(cfg, scanner)
	defer mgr.Cleanup()

	model := tui.NewModel(cfg, mgr, hookEngine, notifier)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err = p.Run()
	return err
}
