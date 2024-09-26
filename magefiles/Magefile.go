// This magefile determines how to build and test the project.
package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/rkennedy/magehelper"
	"github.com/rkennedy/magehelper/tools"
)

// thisDir is the name of the directory, relative to the main module directory, where _this_ module and its go.mod file
// live.
const thisDir = "magefiles"

func goimportsBin() string {
	return filepath.Join("bin", "goimports")
}

func reviveBin() string {
	return filepath.Join("bin", "revive")
}

func logV(s string, args ...any) {
	if mg.Verbose() {
		_, _ = fmt.Printf(s, args...)
	}
}

// Imports formats the code and updates the import statements.
func Imports(ctx context.Context) error {
	mg.CtxDeps(ctx,
		tools.Goimports(goimportsBin()).ModDir(thisDir),
	)
	return nil
}

// Lint performs static analysis on all the code in the project.
func Lint(ctx context.Context) error {
	mg.SerialCtxDeps(ctx,
		Generate,
		tools.Revive(reviveBin(), "revive.toml").ModDir(thisDir),
	)
	return nil
}

// Test runs unit tests.
func Test(ctx context.Context) error {
	return magehelper.Test(ctx)
}

// RunTest runs the specified package's tests.
func RunTest(ctx context.Context, pkg string) error {
	return magehelper.RunTest(ctx, pkg)
}

// BuildTests build all the tests.
func BuildTests(ctx context.Context) error {
	return magehelper.BuildTests(ctx)
}

// Check runs the test and lint targets.
func Check(ctx context.Context) {
	mg.CtxDeps(ctx, Test, Lint)
}

// All runs the build, test, and lint targets.
func All(ctx context.Context) {
	mg.CtxDeps(ctx, Test, Lint)
}

// Generate creates all generated code files.
func Generate(ctx context.Context) {
	mg.CtxDeps(ctx, Imports)
}
