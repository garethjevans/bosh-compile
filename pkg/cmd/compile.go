package cmd

import (
	"bosh-compile/pkg"
	"bosh-compile/pkg/util"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stevenle/topsort"
	"gopkg.in/yaml.v2"
)

type Package struct {
	Name         string   `yaml:"name"`
	Dependencies []string `yaml:"dependencies"`
}

type Manifest struct {
	Name     string    `yaml:"name"`
	Packages []Package `yaml:"packages"`
}

type CompileCmd struct {
	Cmd      *cobra.Command
	Args     []string
	File     string
	Packages []string
}

// NewCompileCmd creates a new compile command.
func NewCompileCmd() *cobra.Command {
	c := &CompileCmd{}

	cmd := &cobra.Command{
		Use:     "compile --file <file> --packages <packages...>",
		Short:   "Compile a bosh release",
		Long:    "",
		Example: "compile --file <file> --packages <packages...>",
		Aliases: []string{"comp"},
		Run: func(cmd *cobra.Command, args []string) {
			c.Cmd = cmd
			c.Args = args
			err := c.Run()
			if err != nil {
				logrus.Fatalf("unable to run command: %s", err)
			}
		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVarP(&c.File, "file", "f", "", "Path to the bosh release to compile")
	cmd.Flags().StringArrayVarP(&c.Packages, "packages", "p", []string{}, "Packages to attempt to compile")

	return cmd
}

func (c *CompileCmd) Run() error {
	if c.File == "" {
		return errors.New("--file parameter is missing")
	}

	if len(c.Packages) == 0 {
		return errors.New("nothing to build, use the --packages parameter")
	}

	tempDir, err := ioutil.TempDir(".", "extracted")
	if err != nil {
		return err
	}

	err = os.MkdirAll(tempDir, 0755)
	if err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	r, err := os.Open(c.File)
	if err != nil {
		return err
	}

	err = pkg.ExtractTarGz(tempDir, r)
	if err != nil {
		return err
	}

	manifest, err := readManifest(tempDir)
	if err != nil {
		return err
	}

	// Initialize the graph.
	graph := topsort.NewGraph()

	for _, p := range manifest.Packages {
		for _, d := range p.Dependencies {
			err = graph.AddEdge(p.Name, d)
			if err != nil {
				return err
			}
		}
	}

	logrus.Infof("Found dependencies")
	for _, p := range manifest.Packages {
		result, err := graph.TopSort(p.Name)
		if err != nil {
			return err
		}
		logrus.Infof("\t%s -> %s", p.Name, result)
	}

	logrus.Infof("Extracting Packages")
	for _, p := range manifest.Packages {
		logrus.Infof("\t%s...", p.Name)
		packageFolder := filepath.Join(tempDir, "packages", p.Name)
		packageGzip := filepath.Join(tempDir, "packages", p.Name+".tgz")
		packageGzipFile, err := os.Open(packageGzip)
		if err != nil {
			return err
		}
		err = pkg.ExtractTarGz(packageFolder, packageGzipFile)
		if err != nil {
			return err
		}
	}

	logrus.Infof("Building Packages\n")

	for _, packageToTest := range c.Packages {
		logrus.Infof("::group::%s", packageToTest)
		logrus.Infof(util.ColorError("###############################################"))
		logrus.Infof("\t\t%s", util.ColorError(packageToTest))
		logrus.Infof(util.ColorError("###############################################"))
		err = BuildAll(tempDir, graph, packageToTest)
		if err != nil {
			logrus.Info("::endgroup::")
			return err
		}
		logrus.Info("::endgroup::")
	}
	return nil
}

func BuildAll(tempDir string, graph *topsort.Graph, packageName string) error {
	// lets start with the db first...
	buildOrder, err := graph.TopSort(packageName)
	if err != nil {
		return err
	}

	for _, build := range buildOrder {
		if isAlreadyBuilt(tempDir, build) {
			logrus.Infof(util.ColorWarn("\t#######################################"))
			logrus.Infof("\t\t%s - Skipping", util.ColorWarn(build))
			logrus.Infof(util.ColorWarn("\t#######################################"))
		} else {
			logrus.Infof(util.ColorError("\t#######################################"))
			logrus.Infof("\t\t%s", util.ColorError(build))
			logrus.Infof(util.ColorError("\t#######################################"))
			workDir := filepath.Join(tempDir, "packages", build)

			boshInstallTarget := filepath.Join(tempDir, "target", build)
			boshInstallTarget, err = filepath.Abs(boshInstallTarget)
			if err != nil {
				return err
			}

			boshCompileTarget, err := filepath.Abs(workDir)
			if err != nil {
				return err
			}

			err = os.MkdirAll(boshInstallTarget, 0755)
			if err != nil {
				return err
			}

			_, err = pkg.Exec(boshCompileTarget, boshInstallTarget, "/bin/bash", "packaging")
			if err != nil {
				logrus.Fatalf("Unable to execute command = %+v", err)
			}

			symlinkPath := fmt.Sprintf("/var/vcap/packages/%s", build)
			logrus.Infof("Creating symlink %s to %s", symlinkPath, boshInstallTarget)

			// remove the symlink if it already exists
			if _, err := os.Lstat(symlinkPath); err == nil {
				if err := os.Remove(symlinkPath); err != nil {
					return fmt.Errorf("failed to unlink: %+v", err)
				}
			}

			// create symbolic link from /var/vcap/packages/<build> -> boshInstallTarget
			err = os.Symlink(boshInstallTarget, symlinkPath)
			if err != nil {
				logrus.Fatalf("Unable to create symlink = %+v", err)
			}
		}
	}
	return nil
}

func isAlreadyBuilt(tempDir string, packageName string) bool {
	boshInstallTarget := filepath.Join(tempDir, "target", packageName)
	if _, err := os.Stat(boshInstallTarget); os.IsNotExist(err) {
		return false
	}
	return true
}

func readManifest(tempDir string) (Manifest, error) {
	manifestPath := filepath.Join(tempDir, "release.MF")
	yamlFile, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return Manifest{}, err
	}

	m := Manifest{}
	err = yaml.Unmarshal(yamlFile, &m)
	if err != nil {
		return Manifest{}, err
	}

	return m, nil
}
