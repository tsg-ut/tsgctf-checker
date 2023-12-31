package checker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Challenge information and test configuration
type Challenge struct {
	Name      string  `json:"name"`
	Timeout   float64 `json:"timeout"`
	Assignee  string  `json:"assignee"`
	SolverDir string
	target    Target
}

type Target struct {
	ChallengeName string
	Host          string
	Port          int
}

// Parse challenge information from a directory.
// The directory must have /solver/info.json file.
// If the challenge name contains spaces, they are replaced with underscores.
func ParseChallenge(path string, targets []Target) (Challenge, error) {
	cfg_file_name := filepath.Join(path, "solver", "info.json")
	cfg_bytes, err := os.ReadFile(cfg_file_name)
	if err != nil {
		return Challenge{}, err
	}

	var chall Challenge
	if err := json.Unmarshal(cfg_bytes, &chall); err != nil {
		return Challenge{}, fmt.Errorf("Failed to parse %s as JSON:\n%v", cfg_file_name, err)
	}

	chall.Name = strings.Replace(chall.Name, " ", "_", -1)
	chall.SolverDir = filepath.Join(path, "solver")

	for _, target := range targets {
		if target.ChallengeName == chall.Name {
			chall.target = target
		}
	}
	if chall.target == (Target{}) {
		return Challenge{}, fmt.Errorf("Target not found for %s", chall.Name)
	}

	return chall, nil
}
