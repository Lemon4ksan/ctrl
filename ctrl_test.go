// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctrl_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/lemon4ksan/ctrl"
	"github.com/lemon4ksan/ctrl/tuikit"
)

func TestCtrl_Facade(t *testing.T) {
	if ctrl.Version == "" {
		t.Fatal("expected non-empty version")
	}

	var executed bool
	cmd := ctrl.NewCommand("test", "test command", "test usage", func(_ context.Context, _ []string, stdout, _ io.Writer) error {
		executed = true
		fmt.Fprintln(stdout, "ok")
		return nil
	})

	app := ctrl.NewApp("testapp", ctrl.Version, "unit test CLI", cmd)
	var stdout, stderr bytes.Buffer
	app.Stdout = &stdout
	app.Stderr = &stderr

	err := app.Run(context.Background(), []string{"test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Fatal("expected command to execute")
	}
	if !strings.Contains(stdout.String(), "ok") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestCtrl_ParseInterspersedFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	v := fs.Bool("v", false, "verbose")
	out := fs.String("o", "", "output")

	args := []string{"input.txt", "-v", "-o=out.bin", "extra.txt"}
	pos, err := ctrl.ParseInterspersedFlags(fs, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !*v {
		t.Errorf("expected -v to be true")
	}
	if *out != "out.bin" {
		t.Errorf("expected -o=out.bin, got %q", *out)
	}
	if len(pos) != 2 || pos[0] != "input.txt" || pos[1] != "extra.txt" {
		t.Errorf("expected [input.txt extra.txt], got %v", pos)
	}
}

func TestCtrl_Widgets(t *testing.T) {
	box := ctrl.NewBox("Card", 30)
	box.AddLine("Line 1")
	str := box.String()
	if !strings.Contains(str, "Card") || !strings.Contains(str, "Line 1") {
		t.Errorf("unexpected box rendering: %s", str)
	}

	tbl := ctrl.NewTable("A", "B")
	tbl.AddRow("1", "2")
	tblStr := tbl.String()
	if !strings.Contains(tblStr, "A") || !strings.Contains(tblStr, "1") {
		t.Errorf("unexpected table rendering: %s", tblStr)
	}

	prog := ctrl.NewProgress(tuikit.WithTotal(100), tuikit.WithInteractive(false))
	prog.SetCurrent(50)
	prog.Done()
}
