package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/RyanTKing/wombats/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	defaultDATS = `(* ****** ****** *)
//
// %s.dats
// Generated by wombats (https://github.com/RyanTKing/wombat)
//
(* ****** ****** *)

#include "./%sstaloadall.hats"

(* ****** ****** *)

implement main0() = () where {
  val _ = print("Hello from %s!\n")
}

(* End of [%s.dats] *)`

	defaultSATS = `(* ****** ****** *)
//
// %s.sats
// Generated by wombats (https://github.com/RyanTKing/wombat)
//
(* ****** ****** *)

(* End of [%s.sats] *)`

	defaultHATS = `(* ****** ****** *)
//
// staloadall.hats
// Generated by wombats (https://github.com/RyanTKing/wombats)
//
(* ****** ****** *)

#staload "./%s%s.sats"

(* End of [staloadall.hats] *)`
)

// newCmd represents the new command
var (
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new Wombats project",
		Long: `Create a new Wombats project in the current directory or in a
specified directory if a name is provided. For example:
	$ wom new     # Initializes a project in the current directory
	$ wom new foo # Creates the directory foo and initializes a project in it`,
		Run: runNew,
	}

	// ErrProjectExists is thrown when a project's directory already exists
	ErrProjectExists = errors.New("project already exists")
)

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringP("name", "n", "",
		"The name of the project (default the directory name)")
	newCmd.Flags().Bool("git", false, "Initialize a new git repository.")
	newCmd.Flags().Bool("lib", false, "Dont create a build directory")
	newCmd.Flags().Bool("cats", false, "Create a CATS directory")
	newCmd.Flags().Bool("small", false,
		"Use a small project template (no DATS/SATS/BUILD dirs")
}

func runNew(cmd *cobra.Command, args []string) {
	small, err := cmd.Flags().GetBool("small")
	if err != nil {
		log.Debug(err)
		log.Fatal("could not check command flag")
	}
	cats, err := cmd.Flags().GetBool("cats")
	if err != nil {
		log.Debug(err)
		log.Fatal("could not check command flag")
	}
	lib, err := cmd.Flags().GetBool("lib")
	if err != nil {
		log.Debug(err)
		log.Fatal("could not check command flag")
	}
	if err := validateNewArgs(small, cats, args); err != nil {
		log.Fatal(err)
	}

	// Make the directory and switch to it
	if err := makeProjectDir(args[0]); err != nil {
		log.Debugln("mkdir error: %s", err)
		log.Fatalf("could not create directory: '%s'", args[0])
	}

	// Get the directory name
	projName, err := getProjName()
	if err != nil {
		log.Debug(err)
		log.Fatal("could not get current directory")
	}

	// Get the project name
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		log.Debug(err)
		log.Fatal("could not check command flag")
	}
	if name == "" {
		name = projName
	}

	// Get the initial config and write it to a file.
	config := config.New(name, projName, small)
	if err := config.Write(); err != nil {
		log.Debug(err)
		log.Fatal("could not create 'Wombats.toml' file")
	}

	git, err := cmd.Flags().GetBool("git")
	if err != nil {
		log.Debug(err)
		log.Fatal("could not check command flag")
	}
	if git {
		if err := initGitRepo(); err != nil {
			log.Fatal(err)
		}
	}

	// Create directories
	if !small {
		if err := createDirs(lib, cats); err != nil {
			log.Debug(err)
			log.Fatal("could not create directories")
		}
	}

	// Create default files
	if err := createDefaultFiles(name, projName, small); err != nil {
		log.Debug(err)
		log.Fatal("could not create default files")
	}

	log.Infof("Created application '%s' project", name)
}

// validateNewArgs validates the arguments given to new
func validateNewArgs(small, cats bool, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no project path provided")
	}
	if len(args) > 1 {
		return fmt.Errorf("found unexpected argument '%s'", args[1])
	}
	if small && cats {
		return fmt.Errorf("can't specify a CATS directory in a small project")
	}

	return nil
}

// makeProjectDir makes a new directory for the project if specified
func makeProjectDir(name string) error {
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return ErrProjectExists
	}

	if err := os.Mkdir(name, os.ModePerm); err != nil {
		return err
	}

	return os.Chdir(name)
}

func getProjName() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Base(wd), nil
}

// initGitRepo initializes the git repo if the option is specified
func initGitRepo() error {
	cmd := exec.Command("git", "init")
	if err := cmd.Run(); err != nil {
		log.Debug(err)
		return err
	}

	return nil
}

func createDirs(lib, cats bool) error {
	dirs := []string{"SATS", "DATS"}
	if !lib {
		dirs = append(dirs, "BUILD")
	}
	if cats {
		dirs = append(dirs, "CATS")
	}
	for _, dir := range dirs {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func createDefaultFiles(name, projName string, small bool) error {
	// Setup the default text and paths for the files
	files := map[string]*struct {
		text string
		path string
		f    *os.File
	}{
		"dats": {},
		"sats": {
			text: fmt.Sprintf(defaultSATS, name, name),
		},
		"hats": {
			path: "staloadall.hats",
		},
	}

	if small {
		files["dats"].text = fmt.Sprintf(defaultDATS, projName, "", name,
			projName)
		files["dats"].path = fmt.Sprintf("%s.dats", projName)
		files["sats"].path = fmt.Sprintf("%s.sats", projName)
		files["hats"].text = fmt.Sprintf(defaultHATS, "", projName)
	} else {
		files["dats"].text = fmt.Sprintf(defaultDATS, projName, "../", name,
			projName)
		files["dats"].path = fmt.Sprintf("DATS/%s.dats", projName)
		files["sats"].path = fmt.Sprintf("SATS/%s.sats", projName)
		files["hats"].text = fmt.Sprintf(defaultHATS, "SATS/", projName)
	}

	// Create the files and write to them
	for _, file := range files {
		f, err := os.Create(file.path)
		if err != nil {
			return err
		}
		file.f = f
		defer file.f.Close()

		if _, err = file.f.WriteString(file.text); err != nil {
			return err
		}
	}

	return nil
}
