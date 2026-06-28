package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	fs := flag.NewFlagSet("embalses-admin", flag.ExitOnError)
	_ = fs.String("cmd", "", "Admin command (migrate, seed, api-key)")
	_ = fs.Bool("help", false, "Show help")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 || (len(os.Args) == 2 && os.Args[1] == "--help") {
		fmt.Println("embalses-admin — administrative CLI")
		fmt.Println()
		fmt.Println("Usage: embalses-admin [options]")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Commands: migrate, seed, api-key, quota")
		os.Exit(0)
	}

	fmt.Println("embalses-admin: stub — not yet implemented")
}
