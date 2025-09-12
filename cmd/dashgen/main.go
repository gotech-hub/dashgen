package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gotech-hub/dashgen/internal/generator"

	"github.com/gotech-hub/dashgen/internal/parser"
)

// Version information (set by build)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

var (
	flagModule  = flag.String("module", "github.com/your-org/app", "go module path of the target project (for imports)")
	flagRoot    = flag.String("root", ".", "target project root (where model/ lives)")
	flagModel   = flag.String("model", "", "single data.go path to parse (optional)")
	flagForce   = flag.Bool("force", false, "overwrite existing files if present")
	flagDryRun  = flag.Bool("dry", false, "print actions without writing files")
	flagVersion = flag.Bool("version", false, "print version information")
)

func main() {
	flag.Parse()

	// Handle version flag
	if *flagVersion {
		fmt.Printf("DashGen %s\n", Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Build time: %s\n", BuildTime)
		return
	}

	var entities []parser.Entity

	if *flagModel != "" {
		e, err := parser.ParseDataGo(*flagModel)
		if err != nil {
			log.Fatalf("parse %s: %v", *flagModel, err)
		}
		entities = append(entities, e...)
	} else {
		matches, err := filepath.Glob(filepath.Join(*flagRoot, "model/**/data.go"))
		if err != nil {
			log.Fatal(err)
		}
		if len(matches) == 0 {
			log.Fatalf("no data.go found under %s", filepath.Join(*flagRoot, "model/**/data.go"))
		}
		for _, p := range matches {
			e, perr := parser.ParseDataGo(p)
			if perr != nil {
				log.Fatalf("parse %s: %v", p, perr)
			}
			entities = append(entities, e...)
		}
	}

	fmt.Printf("Total entities to generate: %d\n", len(entities))
	for i, e := range entities {
		fmt.Printf("Entity %d: %s (pkg: %s, db: %s)\n", i+1, e.Name, e.PkgPath, e.DBName)
	}

	cfg := generator.Config{
		ModulePath:  *flagModule,
		ProjectRoot: *flagRoot,
		Force:       *flagForce,
		DryRun:      *flagDryRun,
	}
	if err := generator.Generate(entities, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "generate error:", err)
		os.Exit(1)
	}
	fmt.Println("✅ Generation finished.")
}
