package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tsg-ut/tsgctf-checker/checker"
	"go.uber.org/zap"
)

func create_conf(logger *zap.SugaredLogger) (checker.CheckerConfig, error) {
	// command-line option overrides configuration of config file.
	conffile := flag.String("config", "config.json", "Configuration file path.")
	retries := flag.Uint("retry", 0, "Number of retries when a test fails.")
	challs_dir := flag.String("challs", "challs", "Challenges directory.")
	parallel := flag.Uint("parallel", 1, "Number of parallel tests.")
	skip_non_exist := flag.Bool("skip-non-exist", false, "Skip challenges who don't have info.json.")
	extra_docker_arg := flag.String("extra-docker-arg", "", "Extra docker arguments passed to \"run\" command.")
	targets_file := flag.String("targets", "targets.json", "Targets file path.")
	notify_slack := flag.Bool("notify-slack", false, "Notify slack when a test fails.")
	dryrun := flag.Bool("dryrun", false, "Dryrun mode. (Don't update database.)")
	target_tests := flag.String("t", "", "Target tests to run.")
	verbose := flag.Bool("verbose", false, "Verbose logging mode.")
	flag.Parse()

	conf, err := checker.ReadConf(*conffile)
	if err != nil {
		return conf, err
	}

	// Override with command-line options
	unknown_flags := make([]string, 0)
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "retries":
			conf.Retries = *retries
			break
		case "parallel":
			conf.ParallelNum = *parallel
			break
		case "challs":
			conf.ChallsDir = *challs_dir
			break
		case "skip-non-exist":
			conf.SkipNonExist = *skip_non_exist
			break
		case "targets":
			conf.TargetsFile = *targets_file
			break
		case "notify-slack":
			if (conf.SlackToken == "" || conf.SlackChannel == "") && *notify_slack {
				logger.Fatal("Slack notification is enabled, but slack_token or slack_channel is not set in config.")
			}
			conf.NotifySlack = *notify_slack
			break
		case "extra-docker-arg":
			conf.ExtraDockerArg = *extra_docker_arg
			break
		case "dryrun":
			conf.Dryrun = *dryrun
			break
		case "verbose":
			conf.Vervose = *verbose
			break
		case "t":
			conf.TargetTests = *target_tests
			break
		case "config":
			break
		default:
			unknown_flags = append(unknown_flags, f.Name)
		}
	})

	if len(unknown_flags) > 0 {
		return conf, fmt.Errorf("Unknown flags: %s", strings.Join(unknown_flags, ", "))
	}

	return conf, nil
}

func main() {
	level := zap.NewAtomicLevel()
	level.SetLevel(zap.DebugLevel)
	slogger, _ := zap.NewDevelopment()
	defer slogger.Sync()
	logger := slogger.Sugar()

	conf, err := create_conf(logger)
	if err != nil {
		logger.Fatal(err)
	}

	var db *sqlx.DB
	if conf.Dryrun == false {
		db, err = checker.Connect(os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBHOST"), os.Getenv("DBNAME"))
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		db = nil
	}

	if err := checker.RunRecordTests(logger, conf, db); err != nil {
		logger.Fatal(err)
	}
}
