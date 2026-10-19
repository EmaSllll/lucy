package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type lockScenario struct {
	// sandbox names the generated server that supplies pinned oracle bytes.
	// Scenarios without one pin their own digests below, so fixtures that are
	// not real servers never need a generated directory or a committed binary.
	sandbox          string
	distribution     string
	minecraft        string
	coreVersion      string
	provider         string
	projectID        string
	filename         string
	bootstrap        string
	coreOutput       string
	coreFixture      string
	minecraftFixture string
	minecraftDigest  string
	requiredFiles    []string
	requiredInputs   []string
	launchArguments  string
	runtimeArtifacts []runtimeArtifactExpectation
	packages         []packageExpectation

	// pending marks a scenario whose core Lucy cannot install yet. The suite
	// prints the blocker and keeps the run green instead of failing it.
	pending string
}

type runtimeArtifactExpectation struct {
	sandboxFile, runtimeFile string
}

type packageExpectation struct {
	reference, projectID, releaseID, fileID, filename, sha256 string
}

var lockScenarios = map[string]lockScenario{
	"vanilla-1201": {
		sandbox:          "vanilla-1201",
		distribution:     "vanilla",
		minecraft:        "1.20.1",
		coreVersion:      "1.20.1",
		provider:         "mojang",
		projectID:        "minecraft",
		filename:         "server.jar",
		coreOutput:       "server.jar",
		coreFixture:      "server.jar",
		minecraftFixture: "server.jar",
	},
	"fabric-1165": {
		distribution:    "fabric",
		minecraft:       "1.16.5",
		coreVersion:     "0.19.5",
		provider:        "fabric",
		projectID:       "fabric-loader",
		filename:        "fabric-loader-0.19.5.jar",
		bootstrap:       "fabric",
		coreOutput:      "libraries/net/fabricmc/fabric-loader/0.19.5/fabric-loader-0.19.5.jar",
		minecraftDigest: "58f329c7d2696526f948470aa6fd0b45545039b64cb75015e64c12194b373da6",
		requiredFiles: []string{
			"run.sh",
			"run.bat",
			"fabric-server-launcher.properties",
		},
	},
	"fabric-118": {
		distribution:    "fabric",
		minecraft:       "1.18",
		coreVersion:     "0.19.5",
		provider:        "fabric",
		projectID:       "fabric-loader",
		filename:        "fabric-loader-0.19.5.jar",
		bootstrap:       "fabric",
		coreOutput:      "libraries/net/fabricmc/fabric-loader/0.19.5/fabric-loader-0.19.5.jar",
		minecraftDigest: "8417edceb471b38a59a6ce260f3aa1479e95264a38eb56283bdf24ec98647e6f",
		requiredFiles: []string{
			"run.sh",
			"run.bat",
			"fabric-server-launcher.properties",
		},
	},
	"fabric-1214": {
		distribution:    "fabric",
		minecraft:       "1.21.4",
		coreVersion:     "0.16.9",
		provider:        "fabric",
		projectID:       "fabric-loader",
		filename:        "fabric-loader-0.16.9.jar",
		bootstrap:       "fabric",
		coreOutput:      "libraries/net/fabricmc/fabric-loader/0.16.9/fabric-loader-0.16.9.jar",
		minecraftDigest: "1066970b09e9c671844572291c4a871cc1ac2b85838bf7004fa0e778e10f1358",
		requiredFiles: []string{
			"run.sh",
			"run.bat",
			"fabric-server-launcher.properties",
		},
		packages: []packageExpectation{
			{
				reference: "modrinth:fabric-api", projectID: "P7dR8mSH", releaseID: "p96k10UR",
				fileID: "9aQezMA2", filename: "fabric-api-0.119.4+1.21.4.jar",
				sha256: "d183bacb845167f09264c2f90322b7ecffe8826debda6f60e597889264bef4af",
			},
			{
				reference: "modrinth:flan", projectID: "Si383TIH", releaseID: "9EC25rAV",
				fileID: "HjsqW7BQ", filename: "flan-1.21.4-1.11.7.b-fabric.jar",
				sha256: "33c466b7c319f1597798f62d13b619317590bf069c4a2c11715fb19fdd98e7f8",
			},
		},
	},
	// Pre-1.18 Forge installers embed no launcher scripts; the forge-installer
	// bootstrap only completes the modern layout, so these scenarios stay
	// pending until legacy installs land.
	"forge-1122": {
		sandbox:      "forge-1122",
		pending:      "pre-1.18 launch layout: the installer embeds no launcher scripts",
		distribution: "forge",
		minecraft:    "1.12.2",
		coreVersion:  "14.23.5.2859",
		provider:     "forge",
		projectID:    "forge",
		filename:     "forge-1.12.2-14.23.5.2859-installer.jar",
		bootstrap:    "forge-installer",
		coreFixture:  "forge-1.12.2-14.23.5.2859-installer.jar",
	},
	"forge-1144": {
		sandbox:      "forge-1144",
		pending:      "pre-1.18 launch layout: the installer embeds no launcher scripts",
		distribution: "forge",
		minecraft:    "1.14.4",
		coreVersion:  "28.2.26",
		provider:     "forge",
		projectID:    "forge",
		filename:     "forge-1.14.4-28.2.26-installer.jar",
		bootstrap:    "forge-installer",
		coreFixture:  "forge-1.14.4-28.2.26-installer.jar",
	},
	"forge-1165": {
		sandbox:      "forge-1165",
		pending:      "pre-1.18 launch layout: the installer embeds no launcher scripts",
		distribution: "forge",
		minecraft:    "1.16.5",
		coreVersion:  "36.2.34",
		provider:     "forge",
		projectID:    "forge",
		filename:     "forge-1.16.5-36.2.34-installer.jar",
		bootstrap:    "forge-installer",
		coreFixture:  "forge-1.16.5-36.2.34-installer.jar",
	},
	"forge-1710": {
		sandbox:      "forge-1710",
		pending:      "legacy installer profile and pre-1.18 launch layout",
		distribution: "forge",
		minecraft:    "1.7.10",
		coreVersion:  "10.13.4.1614",
		provider:     "forge",
		projectID:    "forge",
		filename:     "forge-1.7.10-10.13.4.1614-1.7.10-installer.jar",
		bootstrap:    "forge-installer",
		coreFixture:  "forge-1.7.10-10.13.4.1614-1.7.10-installer.jar",
	},
	"forge-1201": {
		sandbox:         "forge-1201",
		distribution:    "forge",
		minecraft:       "1.20.1",
		coreVersion:     "47.4.10",
		provider:        "forge",
		projectID:       "forge",
		filename:        "forge-1.20.1-47.4.10-installer.jar",
		bootstrap:       "forge-installer",
		coreFixture:     "forge-1.20.1-47.4.10-installer.jar",
		requiredFiles:   []string{"run.sh", "user_jvm_args.txt"},
		requiredInputs:  []string{"libraries/net/minecraftforge/forge/1.20.1-47.4.10/forge-1.20.1-47.4.10-universal.jar"},
		launchArguments: "libraries/net/minecraftforge/forge/1.20.1-47.4.10/unix_args.txt",
		runtimeArtifacts: []runtimeArtifactExpectation{{
			sandboxFile: "libraries/net/minecraftforge/forge/1.20.1-47.4.10/forge-1.20.1-47.4.10-universal.jar",
			runtimeFile: "libraries/net/minecraftforge/forge/1.20.1-47.4.10/forge-1.20.1-47.4.10-universal.jar",
		}},
	},
	"neoforge-211": {
		sandbox:         "neoforge-211",
		distribution:    "neoforge",
		minecraft:       "1.21.1",
		coreVersion:     "21.1.248",
		provider:        "neoforge",
		projectID:       "neoforge",
		filename:        "neoforge-21.1.248-installer.jar",
		bootstrap:       "forge-installer",
		coreFixture:     "neoforge-21.1.248-installer.jar",
		requiredFiles:   []string{"run.sh", "user_jvm_args.txt"},
		requiredInputs:  []string{"libraries/net/neoforged/neoforge/21.1.248/neoforge-21.1.248-universal.jar"},
		launchArguments: "libraries/net/neoforged/neoforge/21.1.248/unix_args.txt",
		runtimeArtifacts: []runtimeArtifactExpectation{{
			sandboxFile: "libraries/net/neoforged/neoforge/21.1.248/neoforge-21.1.248-universal.jar",
			runtimeFile: "libraries/net/neoforged/neoforge/21.1.248/neoforge-21.1.248-universal.jar",
		}},
	},
}

func executeScenario(ctx context.Context, binary, sandboxes, id string) (resultErr error) {
	scenario, ok := lockScenarios[id]
	if !ok {
		return fmt.Errorf("unknown install scenario %q", id)
	}
	if scenario.pending != "" {
		// Pending scenarios are registered but not runnable yet: report the
		// blocker and keep the run green instead of failing on it.
		fmt.Printf("skip %s: %s\n", id, scenario.pending)
		return nil
	}
	oracle := ""
	if scenario.sandbox != "" {
		oracle = filepath.Join(sandboxes, scenario.sandbox)
		info, err := os.Stat(oracle)
		if err != nil {
			return fmt.Errorf("generate %s with task sandboxes:gen first: %w", scenario.sandbox, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("sandbox is not a directory: %s", oracle)
		}
	}
	intent, err := os.ReadFile(filepath.Join("e2e", "lock", "testdata", id+".yaml"))
	if err != nil {
		return fmt.Errorf("read scenario manifest: %w", err)
	}
	var declared manifestIntent
	if err := yaml.Unmarshal(intent, &declared); err != nil {
		return fmt.Errorf("decode scenario intent: %w", err)
	}
	temporary, err := os.MkdirTemp("", "lucy-e2e-install-"+id+"-*")
	if err != nil {
		return err
	}
	defer func() {
		resultErr = errors.Join(resultErr, os.RemoveAll(temporary))
	}()
	root := filepath.Join(temporary, "workspace")
	if err := os.Mkdir(root, 0o755); err != nil {
		return err
	}
	runner := commandRunner{binary: binary, home: filepath.Join(temporary, "home")}
	if err := os.WriteFile(filepath.Join(root, "lucy.yaml"), intent, 0o644); err != nil {
		return err
	}
	if _, err := runner.execute(ctx, root, "install"); err != nil {
		return fmt.Errorf("fresh install: %w", err)
	}
	locked, err := readLock(root)
	if err != nil {
		return err
	}
	if err := scenario.verify(root, oracle, declared, locked); err != nil {
		return fmt.Errorf("installed runtime: %w", err)
	}
	fmt.Printf("install %s: generated lock and runtime outputs\n", id)
	return nil
}
