package checker

import (
	"context"
	"os"
	"testing"
)

func TestChecker_RunRecordTests(t *testing.T) {
	ctx := context.Background()
	container, err := setupMysql(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)
	db, _ := container.OpenDB(ctx)
	logger := create_logger()

	cwd := testing_cd_root(t)
	defer os.Chdir(cwd)

	conf := CheckerConfig{
		ParallelNum:  10,
		ChallsDir:    "tests/assets/challs",
		HaveGenreDir: false,
		Retries:      3,
	}

	// run tests
	if err := RunRecordTests(logger, conf, db); err != nil {
		t.Error(err)
	}

	// check results
	type result struct {
		name   string
		result TestResult
	}
	entries := []result{
		{
			name:   "just-success",
			result: ResultSuccess,
		},
		{
			name:   "just-fail",
			result: ResultFailure,
		},
		{
			name:   "just-success-long",
			result: ResultTimeout,
		},
	}
	for _, ent := range entries {
		results, err := FetchResult(db, ent.name, 1)
		if err != nil {
			t.Error(err)
		}

		if len(results) != 1 {
			t.Errorf("[%s] Expected 1 result, got %d", ent.name, len(results))
		}
		if results[0].Name != ent.name {
			t.Errorf("[%s] Expected name %s, got %s", ent.name, ent.name, results[0].Name)
		}
		if results[0].Result != ent.result {
			t.Errorf("[%s] Expected result %v, got %v", ent.name, ent.result, results[0].Result)
		}
	}
}
