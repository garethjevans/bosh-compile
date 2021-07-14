package pkg

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Exec(workDir string, installTarget string, command string, args ...string) (string, error) {
	log.Printf("cd %s", workDir)
	log.Printf("BOSH_INSTALL_TARGET=%s %s %s", installTarget, command, strings.Join(args, " "))

	path := os.Getenv("PATH")
	binDir, err := filepath.Abs("bin")
	if err != nil {
		return "", err
	}
	path = binDir + ":" + path
	cmd := exec.Command(command, args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	environment := os.Environ()
	environment = append(environment, "BOSH_INSTALL_TARGET="+installTarget)
	environment = append(environment, "BOSH_COMPILE_TARGET="+workDir)
	environment = append(environment, "PATH="+path)
	log.Printf("Configuring environment as %s", environment)
	cmd.Env = environment

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// Print the output
	return "", nil
}
