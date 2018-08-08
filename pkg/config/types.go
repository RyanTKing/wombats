package config

// PackageConfig holds the global package information.
type PackageConfig struct {
	Name       string
	Authors    []string
	Version    string
	License    string
	EntryPoint string
	PatsccArgs []string
	Clibs      []string
	GccArgs    []string
}

// DependencyConfig holds information about a dependency.
type DependencyConfig struct {
	Version string
	Source  string
}

// Config is a struct representing the Wombats.yaml file that each project
// must contain.
type Config struct {
	Package      PackageConfig
	Dependencies []DependencyConfig
}
