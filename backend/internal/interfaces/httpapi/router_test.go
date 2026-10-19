package httpapi

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/application/build"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/internal/domain/logistics"
	"github.com/gin-gonic/gin"
)

type queries struct{ calls int }

func (q *queries) Regions(context.Context) ([]logistics.Region, error) {
	return []logistics.Region{}, nil
}
func (q *queries) RequestBuild(context.Context, string) (build.BuildResult, error) {
	q.calls++
	return build.BuildResult{JobID: "job"}, nil
}
func (q *queries) GetJob(context.Context, string) (build.Job, error) {
	return build.Job{}, build.ErrNotFound
}
func (q *queries) GetGraph(_ context.Context, id string) (logistics.LogisticsGraph, error) {
	if id == "missing" {
		return logistics.LogisticsGraph{}, build.ErrNotFound
	}
	return logistics.LogisticsGraph{ID: id, Nodes: []logistics.LogisticsNode{{ID: "n", Modes: []logistics.TransportMode{logistics.TransportRoad}}, {ID: "port", Modes: []logistics.TransportMode{logistics.TransportSea}}}}, nil
}
func TestAPIValidationAndExports(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := &queries{}
	r := NewRouter(q)
	cases := []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/health", "", 200}, {"POST", "/api/graphs", `{"region_id":"r"}`, 202},
		{"POST", "/api/graphs", `{"region_id":"r","url":"http://example.com"}`, 400},
		{"POST", "/api/graphs", `{"region_id":"r"} {}`, 400}, {"POST", "/api/graphs", `{"region_id":" r"}`, 400},
		{"POST", "/api/graphs", `null`, 400}, {"POST", "/api/graphs", strings.Repeat(" ", 5000), 400},
		{"GET", "/api/jobs/missing", "", 404}, {"GET", "/api/graphs/missing", "", 404},
		{"GET", "/api/graphs/g?mode=invalid", "", 400}, {"GET", "/api/graphs/g/export/geojson", "", 200},
		{"GET", "/api/graphs/g/export/bad", "", 400}, {"GET", "/api/graphs/g/coverage", "", 200},
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(c.method, c.path, strings.NewReader(c.body)))
		if w.Code != c.status {
			t.Errorf("%s %s => %d: %s", c.method, c.path, w.Code, w.Body.String())
		}
	}
	if q.calls != 1 {
		t.Fatalf("invalid request reached application: %d calls", q.calls)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/graphs/g?mode=road", nil))
	if strings.Contains(w.Body.String(), `"id":"port"`) {
		t.Fatal("mode filter leaked sea node")
	}
}
