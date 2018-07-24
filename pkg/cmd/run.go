package cmd

import (
	"os"
	"os/exec"

	"github.com/RyanTKing/wombats/pkg/ats"
	"github.com/RyanTKing/wombats/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the current project",
	Long: `Compile the project if necessary then run it if successfully built.
	
All arguments passed to the run command will be passed to patscc.`,
	Run: runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) {
	config, err := config.Read()
	if err != nil {
		log.Debugf("error reading config: %s", err)
		log.Fatalf(
			"could not find '%s' in this directory or any parent directory",
			"Wombats.toml",
		)
	}

	projName, err := getProjName()
	if err != nil {
		log.Fatalf("could not build '%s' project", config.Package.Name)
	}

	execFile := ats.Build(projName, config.Package.EntryPoint)
	log.Infof("Running '%s'", execFile)
	execCmd := exec.Command(execFile, args...)
	execCmd.Env = os.Environ()
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	if err := execCmd.Run(); err != nil {
		log.Debugf("error running '%s': %s", config.Package.Name, err)
		log.Fatalf("could not run '%s' project", config.Package.Name)
	}

}
