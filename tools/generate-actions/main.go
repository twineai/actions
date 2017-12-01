package main

import (
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/namsral/flag"
)

func write(t *template.Template, path string, data interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	return t.Execute(f, data)
}

func main() {
	flag.Parse()

	// Don't print any sort of timestamp information.
	log.SetFlags(0)

	if len(FlagActionDir) == 0 {
		log.Fatal("must provide an action dir via --action-dir")
	}

	actions := flag.Args()
	if len(actions) == 0 {
		log.Fatal("must provide at least one action name")
	}

	packageTemplate, err := template.New("package").Parse(packageTemplateStr)
	if err != nil {
		log.Fatalf("error creating package template: %v", err)
	}

	indexTemplate, err := template.New("index").Parse(indexTemplateStr)
	if err != nil {
		log.Fatalf("error creating index template: %v", err)
	}

	for _, actionName := range actions {
		log.Printf("Creating action: %s", actionName)

		actionPath := filepath.Join(FlagActionDir, actionName)

		_, err := os.Stat(actionPath)
		if !os.IsNotExist(err) {
			if err != nil {
				log.Fatalf("error stating action dir %s: %v", actionPath, err)
			} else {
				log.Fatalf("action directory already exists: %s", actionPath)
			}
		}

		if err := os.MkdirAll(actionPath, 0755); err != nil {
			log.Fatalf("error creating action dir %s: %v", actionPath, err)
		}

		actionData := struct {
			ActionName string
		}{
			ActionName: actionName,
		}

		packagePath := filepath.Join(actionPath, "package.json")
		err = write(packageTemplate, packagePath, actionData)
		if err != nil {
			log.Fatalf("error writing package file: %v", err)
		}

		indexPath := filepath.Join(actionPath, "index.js")
		err = write(indexTemplate, indexPath, actionData)
		if err != nil {
			log.Fatalf("error writing index file: %v", err)
		}
	}
}
