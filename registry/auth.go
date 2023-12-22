package registry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/registry"
)

type DockerConfig struct {
	Auths map[string]any `json:"auths"`
	Store string         `json:"credsStore"`
}

func GetEncodedAuth(ctx context.Context, image string) string {
	authConfig := registry.AuthConfig{}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("could not get home dir: %v", err)
		return ""
	}

	configBytes, err := os.ReadFile(filepath.Join(home, ".docker", "config.json"))
	if err != nil {
		log.Printf("could not read docker config: %v", err)
		return ""
	}

	var cfg DockerConfig
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		log.Printf(".docker/config.json is invalid JSON: %v", err)
		return ""
	}

	imageHost := strings.Split(image, "/")[0]
	var hostAuth any
	for host, savedAuth := range cfg.Auths {
		if host == imageHost {
			hostAuth = savedAuth
			break
		}

		if u, err := url.Parse(host); err == nil {
			if host == u.Host {
				hostAuth = savedAuth
				break
			}
		}
	}

	if hostAuth == nil {
		var ok bool
		imageHost = "index.docker.io"
		if hostAuth, ok = cfg.Auths["https://index.docker.io/v1"]; !ok {
			return ""
		}
	}

	if authStr, ok := hostAuth.(string); ok {
		userPass, err := base64.StdEncoding.DecodeString(authStr)
		if err != nil {
			return ""
		}

		parts := strings.Split(string(userPass), ":")
		authConfig.Username = parts[0]
		authConfig.Password = parts[1]
	} else {
		output := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, fmt.Sprintf("docker-credential-%s", cfg.Store), "get")
		cmd.Stdin = strings.NewReader(imageHost)
		cmd.Stdout = output
		if err = cmd.Run(); err != nil {
			log.Printf("could not get creds from credsStore %s: %v", cfg.Store, err)
			return ""
		}

		var creds map[string]string
		if err = json.Unmarshal(output.Bytes(), &creds); err != nil {
			log.Printf("could not parse creds json: %v", err)
			return ""
		}

		authConfig.Username = creds["Username"]
		if p, ok := creds["Password"]; ok {
			authConfig.Password = p
		}

		if i, ok := creds["Secret"]; ok {
			authConfig.IdentityToken = i
		}

		if s, ok := creds["ServerURL"]; ok {
			authConfig.ServerAddress = s
		}
	}

	ac, err := registry.EncodeAuthConfig(authConfig)
	if err != nil {
		log.Printf("could not encode auth config: %v", err)
		return ""
	}

	return ac
}
