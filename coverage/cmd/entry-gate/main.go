package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/J-S-Te/license-core/coverage"
	"os"
)

func main() {
	root := flag.String("root", "../../..", "workspace root")
	registry := flag.String("registry", ".ai/qa/commercial-license/entry-registry.json", "registry relative to workspace root")
	generate := flag.Bool("inventory", false, "write inventory JSON to stdout (no implicit classification)")
	output := flag.String("inventory-output", "", "optional generated inventory file (requires -inventory)")
	flag.Parse()
	current, err := coverage.Scan(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *generate {
		if *output != "" {
			raw, marshalErr := json.MarshalIndent(current, "", "  ")
			if marshalErr == nil {
				marshalErr = os.WriteFile(*output, append(raw, '\n'), 0644)
			}
			if marshalErr != nil {
				fmt.Fprintln(os.Stderr, marshalErr)
				os.Exit(1)
			}
			return
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err = enc.Encode(current); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *output != "" {
		fmt.Fprintln(os.Stderr, "inventory-output requires inventory")
		os.Exit(1)
	}
	raw, err := os.ReadFile(*root + "/" + *registry)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var saved coverage.Registry
	if err = json.Unmarshal(raw, &saved); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = coverage.Check(current, saved); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("source registry passed: %d entries; runtime route and activity registration tests remain separately required\n", len(saved.Entries))
}
