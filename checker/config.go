package checker

import (
	"encoding/json"
	"os"
)

type CheckerConfig struct {
	ParallelNum  uint   `json:"parallel"`
	ChallsDir    string `json:"challs_dir"`
	Retries      uint   `json:"retries"`
	SkipNonExist bool   `json:"skip_non_exist"`
	SlackToken   string `json:"slack_token"`
	SlackChannel string `json:"slack_channel"`
	NotifySlack  bool
}

func ReadConf(config_path string) (CheckerConfig, error) {
	cfg_bytes, err := os.ReadFile(config_path)
	if err != nil {
		return CheckerConfig{}, err
	}

	var conf CheckerConfig
	if err := json.Unmarshal(cfg_bytes, &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
