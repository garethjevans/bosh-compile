package cmd

import (
	"bosh-compile/pkg"
	"io/ioutil"
	"log"
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
	tempDir, err := ioutil.TempDir(".", "extracted")
	if err != nil {
		return err
	}
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	r, err := os.Open(c.File)
	if err != nil {
		return err
	}
	pkg.ExtractTarGz(tempDir, r)

	manifest, err := readManifest(tempDir)
	if err != nil {
		return err
	}

	// Initialize the graph.
	graph := topsort.NewGraph()

	for _, p := range manifest.Packages {
		for _, d := range p.Dependencies {
			graph.AddEdge(p.Name, d)
		}
	}

	log.Printf("Found dependencies\n")
	for _, p := range manifest.Packages {
		result, err := graph.TopSort(p.Name)
		if err != nil {
			return err
		}
		log.Printf("\t%s -> %s\n", p.Name, result)
	}

	log.Printf("Extracting Packages\n")
	for _, p := range manifest.Packages {
		log.Printf("\t%s...\n", p.Name)
		packageFolder := filepath.Join(tempDir, "packages", p.Name)
		packageGzip := filepath.Join(tempDir, "packages", p.Name+".tgz")
		packageGzipFile, err := os.Open(packageGzip)
		if err != nil {
			panic(err)
		}
		pkg.ExtractTarGz(packageFolder, packageGzipFile)
	}

	for _, packageToTest := range c.Packages {
		err = BuildAll(tempDir, graph, packageToTest)
		if err != nil {
			return err
		}
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
			log.Println("build > " + build + ", Skipping")
		} else {
			log.Println("build > " + build)
			workDir := filepath.Join(tempDir, "packages", build)

			boshInstallTarget := filepath.Join(tempDir, "target", build)
			boshInstallTarget, err = filepath.Abs(boshInstallTarget)
			if err != nil {
				return err
			}
			os.MkdirAll(boshInstallTarget, 0755)
			output, err := pkg.Exec(workDir, boshInstallTarget, "/bin/bash", "packaging")
			log.Println(output)
			if err != nil {
				log.Fatalf("Unable to execute command = %s, %s", err, output)
				return err
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
