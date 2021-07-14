package pkg

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func ExtractTarGz(tempDir string, gzipStream io.Reader) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			dir := filepath.Join(tempDir, header.Name)
			if err := os.Mkdir(dir, 0755); err != nil {
				panic(err)
			}
		case tar.TypeReg:
			file := filepath.Join(tempDir, header.Name)
			outFile, err := os.Create(file)
			if err != nil {
				panic(err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				panic(err)
			}
			outFile.Close()

		default:
			panic(fmt.Sprintf(
				"ExtractTarGz: uknown type: %s in %s",
				header.Typeflag,
				header.Name))
		}

	}
}
