package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	fs := flag.NewFlagSet("embalses-ingest", flag.ExitOnError)
	_ = fs.String("source", "", "Source to ingest (e.g. miteco, snczi, ign)")
	_ = fs.String("since", "", "Ingest data since date (YYYY-MM-DD)")
	_ = fs.Bool("dry-run", false, "Parse and validate without writing to DB")
	_ = fs.Bool("help", false, "Show help")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 || (len(os.Args) == 2 && os.Args[1] == "--help") {
		fmt.Println("embalses-ingest — ingestion CLI for reservoir data")
		fmt.Println()
		fmt.Println("Usage: embalses-ingest [options]")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Sources: miteco, snczi, ign, saih-ebro, saih-jucar, aemet")
		os.Exit(0)
	}

	fmt.Println("embalses-ingest: stub — not yet implemented")
}
