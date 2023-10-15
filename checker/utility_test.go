package checker

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func create_logger() *zap.SugaredLogger {
	level := zap.NewAtomicLevel()
	level.SetLevel(zap.DebugLevel)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	return logger.Sugar()
}

func testing_cd_root(t *testing.T) string {
	cwd, _ := os.Getwd()

	if err := os.Chdir(".."); err != nil {
		t.Error("Failed to change working directory")
	}

	return cwd
}
