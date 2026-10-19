package logistics_repository_test

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/repository/logistics_repository"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/repository/osm"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/service/graph_build_service"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/service/logistics_service"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/tests/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLogisticsRepositoryBuildLifecycle(t *testing.T) {

	url := os.Getenv("TEST_DATABASE_URL")

	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer cancel()

	admin, err := pgxpool.New(ctx, url)

	if err != nil {
		t.Fatal(err)
	}

	defer admin.Close()

	schema := "logistics_test_" + rand.Text()
	quoted := pgx.Identifier{schema}.Sanitize()

	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}

	defer func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+quoted+" CASCADE")
	}()

	cfg, err := pgxpool.ParseConfig(url)

	if err != nil {
		t.Fatal(err)
	}

	cfg.ConnConfig.RuntimeParams["search_path"] = quoted
	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		t.Fatal(err)
	}

	defer pool.Close()

	repo := logistics_repository.New(pool)
	region := logistics.Region{ID: "test", Name: "Test", Bounds: logistics.Bounds{West: 29, South: 59, East: 31, North: 61}}
	second := region
	second.ID = "second"
	release, err := repo.Initialize(ctx, []logistics.Region{region, second})

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if release != nil {
			release()
		}
	}()

	if unlock, err := repo.Initialize(ctx, []logistics.Region{region}); err == nil {
		unlock()
		t.Fatal("second server obtained exclusive worker lock")
	}

	service := logistics_service.NewService(repo)
	results := make(chan logistics_service.BuildResult, 16)
	failures := make(chan error, 16)
	var wg sync.WaitGroup

	for range 16 {
		wg.Go(func() {
			r, err := service.RequestBuild(ctx, "test")

			if err != nil {
				failures <- err

				return
			}

			results <- r
		})
	}

	wg.Wait()
	close(results)
	close(failures)

	for err := range failures {
		t.Fatal(err)
	}

	id := ""

	for r := range results {
		if id == "" {
			id = r.JobID
		}

		if r.JobID != id || r.GraphID != "" {
			t.Fatalf("request was not deduplicated: %+v", r)
		}
	}

	if _, err := service.RequestBuild(ctx, "unknown"); !errors.Is(err, logistics_service.ErrNotFound) {
		t.Fatalf("unknown region: %v", err)
	}

	root := testutil.FixtureDir(t)
	pipeline := graph_build_service.Service{Policy: logistics.DefaultFreightPolicy(), Source: osmdata.Source{Catalog: osmdata.Catalog{MaximumObjects: 1000, Regions: map[string]osmdata.Files{"test": {OSM: filepath.Join(root, "freight.osm"), References: filepath.Join(root, "references.json")}}}}}
	manager := logistics_service.WorkerService{Repository: repo, Builder: pipeline, Workers: 2, PollInterval: time.Millisecond * 10, BuildTimeout: time.Second * 5}
	workerCtx, stop := context.WithCancel(ctx)
	done := make(chan error, 1)

	go func() {
		done <- manager.Run(workerCtx)
	}()

	defer stop()

	var job logistics_service.Job
	ticker := time.NewTicker(10 * time.Millisecond)

	defer ticker.Stop()

	for {
		job, err = service.GetJob(ctx, id)

		if err != nil {
			t.Fatal(err)
		}

		if job.Status == logistics_service.Completed {
			break
		}

		if job.Status == logistics_service.Failed {
			t.Fatalf("build failed: %+v", job)
		}

		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
	}

	stop()

	if err := <-done; err != nil {
		t.Fatal(err)
	}

	g, err := service.GetGraph(ctx, job.GraphID)

	if err != nil {
		t.Fatal(err)
	}

	if len(g.Nodes) != 3 || len(g.Edges) != 1 {
		t.Fatalf("persisted graph: %+v", g)
	}

	cached, err := service.RequestBuild(ctx, "test")

	if err != nil || cached.GraphID != job.GraphID || cached.JobID != "" {
		t.Fatalf("cache=%+v error=%v", cached, err)
	}

	// Arbitrary areas are immutable, deduplicated and excluded from presets.
	source := pipeline.Source.(osmdata.Source)
	files := source.Catalog.Regions["test"]
	files.Bounds = &region.Bounds
	source.Catalog.Regions["test"] = files
	areaService := logistics_service.NewService(repo, source)
	areaBounds := logistics.Bounds{West: 29.9, South: 59.9, East: 30.1, North: 60.1}
	areaResults := make(chan logistics_service.BuildResult, 8)
	areaFailures := make(chan error, 8)

	for range 8 {
		wg.Go(func() {
			result, err := areaService.RequestAreaBuild(ctx, areaBounds)

			if err != nil {
				areaFailures <- err

				return
			}

			areaResults <- result
		})
	}

	wg.Wait()
	close(areaResults)
	close(areaFailures)

	for err := range areaFailures {
		t.Fatal(err)
	}

	areaJob := ""

	for result := range areaResults {
		if areaJob == "" {
			areaJob = result.JobID
		}

		if result.JobID != areaJob || result.Region.Bounds != areaBounds {
			t.Fatalf("area not deduplicated: %+v", result)
		}
	}

	areaBounds.East = 30.2
	otherArea, err := areaService.RequestAreaBuild(ctx, areaBounds)

	if err != nil {
		t.Fatal(err)
	}

	if otherArea.JobID == areaJob {
		t.Fatal("different areas share task")
	}

	presets, err := repo.Regions(ctx)

	if err != nil || len(presets) != 2 {
		t.Fatalf("generated areas leaked into presets: %v %v", presets, err)
	}

	areaCtx, areaStop := context.WithCancel(ctx)
	areaDone := make(chan error, 1)

	go func() {
		areaDone <- manager.Run(areaCtx)
	}()

	for _, jobID := range []string{areaJob, otherArea.JobID} {
		for {
			job, err := service.GetJob(ctx, jobID)

			if err != nil {
				areaStop()
				t.Fatal(err)
			}

			if job.Status == logistics_service.Completed {
				break
			}

			if job.Status == logistics_service.Failed {
				areaStop()
				t.Fatalf("area failed: %+v", job)
			}

			select {
			case <-ctx.Done():
				areaStop()
				t.Fatal(ctx.Err())
			case <-ticker.C:
			}
		}
	}

	areaStop()

	if err := <-areaDone; err != nil {
		t.Fatal(err)
	}

	cachedArea, err := areaService.RequestAreaBuild(ctx, areaBounds)

	if err != nil || cachedArea.GraphID == "" {
		t.Fatalf("area cache: %+v %v", cachedArea, err)
	}

	result, err := service.RequestBuild(ctx, "second")

	if err != nil {
		t.Fatal(err)
	}

	claimed, _, err := repo.Claim(ctx)

	if err != nil || claimed.ID != result.JobID {
		t.Fatalf("claim=%+v err=%v", claimed, err)
	}

	if _, _, err := repo.Claim(ctx); !errors.Is(err, logistics_service.ErrNotFound) {
		t.Fatalf("claimed same job twice: %v", err)
	}

	release()
	release = nil
	release, err = repo.Initialize(ctx, []logistics.Region{region, second})

	if err != nil {
		t.Fatal(err)
	}

	recovered, err := repo.GetJob(ctx, result.JobID)

	if err != nil || recovered.Status != logistics_service.Pending {
		t.Fatalf("restart did not recover job: %+v %v", recovered, err)
	}

	claimed, _, err = repo.Claim(ctx)

	if err != nil {
		t.Fatal(err)
	}

	if err := repo.Complete(ctx, claimed, g); err == nil {
		t.Fatal("cross-region graph accepted")
	}

	if err := repo.Fail(ctx, claimed.ID, "test failure"); err != nil {
		t.Fatal(err)
	}

	retry, err := repo.RequestBuild(ctx, "second")

	if err != nil || retry.JobID == claimed.ID {
		t.Fatalf("failure retry=%+v %v", retry, err)
	}
}
