package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/sst/sst/v3/cmd/sst/cli"
	"github.com/sst/sst/v3/internal/util"
	"github.com/sst/sst/v3/pkg/npm"
	"github.com/sst/sst/v3/pkg/process"
	"github.com/sst/sst/v3/pkg/project"
)

func CmdInit(cli *cli.Cli) error {
	if _, err := os.Stat("sst.config.ts"); err == nil {
		color.New(color.FgRed, color.Bold).Print("×")
		color.New(color.FgWhite, color.Bold).Println("  SST project already exists")
		return nil
	}

	logo := []string{
		``,
		`   ███████╗███████╗████████╗`,
		`   ██╔════╝██╔════╝╚══██╔══╝`,
		`   ███████╗███████╗   ██║   `,
		`   ╚════██║╚════██║   ██║   `,
		`   ███████║███████║   ██║   `,
		`   ╚══════╝╚══════╝   ╚═╝   `,
		``,
	}

	fmt.Print("\033[?25l")
	// orange
	fmt.Print("\033[38;2;255;127;0m")
	for _, line := range logo {
		for _, char := range line {
			fmt.Print(string(char))
			time.Sleep(5 * time.Millisecond)
		}
		fmt.Println()
	}
	fmt.Print("\033[?25h")

	var template string

	hints := []string{}
	files, err := os.ReadDir(".")
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		hints = append(hints, file.Name())
	}

	color.New(color.FgBlue, color.Bold).Print(">")
	switch {
	case slices.ContainsFunc(hints, func(s string) bool { return strings.HasPrefix(s, "next.config") }):
		fmt.Println("  Next.js detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - modify the tsconfig.json")
		fmt.Println("   - add sst to package.json")
		template = "nextjs"
		break

	case slices.ContainsFunc(hints, func(s string) bool { return strings.HasPrefix(s, "react-router.config") }):
		fmt.Println("  React Router detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - modify the tsconfig.json")
		fmt.Println("   - add sst to package.json")
		template = "react-router"
		break

	case slices.ContainsFunc(hints, func(s string) bool { return strings.HasPrefix(s, "astro.config") }):
		fmt.Println("  Astro detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - modify the tsconfig.json")
		fmt.Println("   - add sst to package.json")
		template = "astro"
		break

	case slices.ContainsFunc(hints, func(s string) bool {
		return strings.HasPrefix(s, "app.config") && fileContains(s, "@solidjs/start")
	}):
		fmt.Println("  SolidStart detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - add sst to package.json")
		template = "solid-start"
		break

	case slices.ContainsFunc(hints, func(s string) bool {
		return strings.HasPrefix(s, "app.config") && fileContains(s, "@tanstack/")
	}):
		fmt.Println("  TanStack Start detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - add sst to package.json")
		template = "tanstack-start"
		break

	case slices.ContainsFunc(hints, func(s string) bool { return strings.HasPrefix(s, "nuxt.config") }):
		fmt.Println("  Nuxt detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - add sst to package.json")
		template = "nuxt"
		break

	case slices.ContainsFunc(hints, func(s string) bool { return strings.HasPrefix(s, "svelte.config") }):
		fmt.Println("  SvelteKit detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - add sst to package.json")
		template = "svelte-kit"
		break

	case slices.ContainsFunc(hints, func(s string) bool {
		return strings.HasPrefix(s, "remix.config") ||
			(strings.HasPrefix(s, "vite.config") && fileContains(s, "@remix-run/dev"))
	}):
		fmt.Println("  Remix detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - add sst to package.json")
		template = "remix"
		break

	case slices.ContainsFunc(hints, func(s string) bool {
		return (strings.HasPrefix(s, "vite.config") && fileContains(s, "@analogjs/platform"))
	}):
		fmt.Println("  Analog detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - add sst to package.json")
		template = "analog"
		break

	case slices.ContainsFunc(hints, func(s string) bool { return strings.HasPrefix(s, "angular.json") }):
		fmt.Println("  Angular detected. This will...")
		fmt.Println("   - create an sst.config.ts")
		fmt.Println("   - add sst to package.json")
		template = "angular"
		break

	case slices.Contains(hints, "package.json"):
		fmt.Println("  JS project detected. This will...")
		fmt.Println("   - use the JS template")
		fmt.Println("   - create an sst.config.ts")
		template = "js"
		break

	default:
		fmt.Println("  No frontend detected. This will...")
		fmt.Println("   - use the vanilla template")
		fmt.Println("   - create an sst.config.ts")
		template = "vanilla"
		break
	}
	fmt.Println()

	p := promptui.Select{
		Items:        []string{"Yes", "No"},
		Label:        "‏‏‎ ‎Continue",
		HideSelected: true,
		HideHelp:     true,
	}

	if !cli.Bool("yes") {
		_, confirm, err := p.Run()
		if err != nil {
			return util.NewReadableError(err, "")
		}
		if confirm == "No" {
			return nil
		}
	}

	color.New(color.FgGreen, color.Bold).Print("✓")
	color.New(color.FgWhite).Println("  Template:", template)
	fmt.Println()

	home := "aws"
	if template == "vanilla" || template == "js" {
		p = promptui.Select{
			Label:        "‏‏‎ ‎Where do you want to deploy your app? You can change this later",
			HideSelected: true,
			Items:        []string{"aws", "cloudflare"},
			HideHelp:     true,
		}
		_, home, err = p.Run()
		if err != nil {
			return util.NewReadableError(err, "")
		}
		color.New(color.FgGreen, color.Bold).Print("✓")
		color.New(color.FgWhite).Println("  Using: " + home)
		fmt.Println()
	}

	if template == "js" {
		template = "js-" + home
	}

	instructions, err := project.Create(template, home)
	if err != nil {
		return err
	}
	var cmd *exec.Cmd

	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = "  Installing providers..."
	spin.Start()

	cfgPath, err := cli.Discover()
	if err != nil {
		return err
	}
	proj, err := project.New(&project.ProjectConfig{
		Config:  cfgPath,
		Stage:   "sst",
		Version: version,
	})
	if err != nil {
		return err
	}
	if err := proj.CopyPlatform(version); err != nil {
		return err
	}

	if err := proj.Install(); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	mgr, _ := npm.DetectPackageManager(cwd)
	if mgr != "" {
		cmd = process.Command(mgr, "install")
		spin.Suffix = "  Installing dependencies..."
		spin.Start()
		slog.Info("installing deps", "args", cmd.Args)
		cmd.Run()
		spin.Stop()
	}

	if template == "nextjs" {
		hasEslint := slices.ContainsFunc(hints, func(s string) bool {
			return strings.Contains(strings.ToLower(s), "eslint")
		})

		if hasEslint {
			configFile := "sst.config.ts"
			content, err := os.ReadFile(configFile)
			if err != nil {
				return err
			}

			newContent := "// eslint-disable-next-line @typescript-eslint/triple-slash-reference\n" + string(content)
			err = os.WriteFile(configFile, []byte(newContent), 0644)
			if err != nil {
				return err
			}
		}
	}

	spin.Stop()

	if len(instructions) == 0 {
		color.New(color.FgGreen, color.Bold).Print("✓")
		color.New(color.FgWhite).Println("  Done 🎉")
	}
	if len(instructions) > 0 {
		for i, instruction := range instructions {
			if i == 0 {
				color.New(color.FgBlue, color.Bold).Print(">")
				fmt.Println("  " + instruction)
				continue
			}
			color.New(color.FgWhite).Println("   " + instruction)
		}
	}
	fmt.Println()
	return nil
}

func fileContains(filePath string, str string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), str) {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		return false
	}

	return false
}
