package libcnbext

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
)

func Run(ctx context.Context, detect DetectFunc, generate GenerateFunc) {
	processName := filepath.Base(os.Args[0])
	switch processName {
	case "detect":
		result, err := detect()
		if err != nil {
			os.Exit(1)
		}

		outputFileName := os.Getenv("CNB_BUILD_PLAN_PATH")
		if outputFileName == "" {
			os.Exit(2)
		}

		outputFile, err := os.Create(outputFileName)
		if err != nil {
			os.Exit(4)
		}

		if err = toml.NewEncoder(outputFile).Encode(result); err != nil {
			os.Exit(8)
		}

		os.Exit(0)
	case "generate":
		buildDockerfile, runDockerfile, err := generate(ctx)
		if err != nil {
			os.Exit(1)
		}

		validate, err := strconv.ParseBool(os.Getenv("BP_EXT_VALIDATE_GENERATED_IMAGES"))
		if err != nil {
			validate = false
		}

		outputDir := os.Getenv("CNB_OUTPUT_DIR")
		if outputDir == "" {
			os.Exit(8)
		}

		if buildDockerfile != "" {
			if validate {
				if err := buildDockerfile.Validate(ctx); err != nil {
					os.Exit(2)
				}
			}

			if err := os.WriteFile(filepath.Join(outputDir, "build.Dockerfile"), []byte(buildDockerfile), 0o644); err != nil {
				os.Exit(16)
			}
		}

		if runDockerfile != "" {
			if validate {
				if err := runDockerfile.Validate(ctx); err != nil {
					os.Exit(4)
				}
			}

			if err := os.WriteFile(filepath.Join(outputDir, "run.Dockerfile"), []byte(runDockerfile), 0o644); err != nil {
				os.Exit(32)
			}
		}
	default:
		log.Fatal("Command must be called 'detect' or 'generate'")
	}
}
