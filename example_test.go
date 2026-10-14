// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctrl_test

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lemon4ksan/ctrl"
)

// Example demonstrates constructing a CLI application using ctrl.
func Example() {
	app := ctrl.NewApp("example", "1.0.0", "Minimal example CLI",
		ctrl.NewCommand("greet", "Print greeting message", "example greet [name]",
			func(ctx context.Context, args []string, stdout, stderr io.Writer) error {
				name := "World"
				if len(args) > 0 {
					name = args[0]
				}
				fmt.Fprintf(stdout, "Hello, %s!\n", name)
				return nil
			},
		),
	)

	_ = app.Run(context.Background(), []string{"greet", "Gopher"})
	// Output:
	// Hello, Gopher!
}

// ExampleParseInterspersedFlags demonstrates parsing POSIX flags and positional arguments.
func ExampleParseInterspersedFlags() {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	v := fs.Bool("v", false, "verbose")
	out := fs.String("o", "", "output binary")

	args := []string{"main.go", "-v", "-o=app.exe", "extra.go"}
	posArgs, _ := ctrl.ParseInterspersedFlags(fs, args)

	fmt.Printf("Verbose: %t\n", *v)
	fmt.Printf("Output: %s\n", *out)
	fmt.Printf("Positionals: %v\n", posArgs)
	// Output:
	// Verbose: true
	// Output: app.exe
	// Positionals: [main.go extra.go]
}

// ExampleNewBox demonstrates creating an auto-sized bordered card panel.
func ExampleNewBox() {
	card := ctrl.NewBox("Status Card", 28)
	card.AddLine("Engine: OK")
	_ = card.Render(os.Stdout)
}

// ExampleNewTable demonstrates creating an auto-aligned data table.
func ExampleNewTable() {
	tbl := ctrl.NewTable("Method", "Route", "Latency")
	tbl.AddRow("GET", "/health", "1.2 ms")
	_ = tbl.Render(os.Stdout)
}
