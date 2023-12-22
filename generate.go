package libcnbext

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gideaworx/libcnbext/registry"
)

var runBaseImage = regexp.MustCompile(`(?i)^\s*FROM\s+(?P<imageURI>\S+)\s*$`)

const (
	buildBaseArg       = "arg base_image"
	buildBaseFrom      = "from ${base_image}"
	buildUserArg       = "arg user_id"
	buildUserDirective = "user ${user_id}"
)

type GenerateFunc func(ctx context.Context) (BuildDockerfile, RunDockerfile, error)

type BuildDockerfile string

func (b BuildDockerfile) Validate(ctx context.Context) error {
	lines := strings.Split(string(b), "\n")

	var (
		hasBaseArg       bool
		hasBaseFrom      bool
		hasUserArg       bool
		hasUserDirective bool
	)
	for _, line := range lines {
		normalized := strings.TrimSpace(strings.ToLower(line))
		switch normalized {
		case buildBaseArg:
			hasBaseArg = true
		case buildBaseFrom:
			hasBaseFrom = true
		case buildUserArg:
			hasUserArg = true
		case buildUserDirective:
			hasUserDirective = true
		}
	}

	if hasBaseArg && hasBaseFrom && hasUserArg && hasUserDirective {
		return nil
	}

	return fmt.Errorf("dockerfile must contain the following lines: %q, %q, %q, %q",
		buildBaseArg, buildBaseFrom, buildUserArg, buildUserDirective)
}

type RunDockerfile string

func (r RunDockerfile) Validate(ctx context.Context) error {
	lines := strings.Split(string(r), "\n")

	fromLine := ""
	for _, line := range lines {
		if runBaseImage.MatchString(line) {
			fromLine = line
			break
		}
	}

	if fromLine == "" {
		return errors.New("no FROM line found in Dockerfile")
	}

	parts := runBaseImage.FindStringSubmatch(fromLine)
	if len(parts) == 2 {
		imageURI := parts[1]
		_, err := registry.ImageExists(ctx, imageURI)
		return err
	}

	return errors.New("invalid FROM line in Dockerfile")
}
