package httpapi_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/interfaces/httpapi"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/service/logistics_service"
	"github.com/gin-gonic/gin"
)

type queries struct {
	calls   int
	areaErr error
}

func (q *queries) Regions(context.Context) ([]logistics.Region, error) {
	return []logistics.Region{}, nil
}

func (q *queries) RequestBuild(context.Context, string) (logistics_service.BuildResult, error) {

	q.calls++

	return logistics_service.BuildResult{JobID: "job"}, nil
}

func (q *queries) RequestAreaBuild(_ context.Context, b logistics.Bounds) (logistics_service.BuildResult, error) {

	q.calls++

	if q.areaErr != nil {
		return logistics_service.BuildResult{}, q.areaErr
	}

	return logistics_service.BuildResult{JobID: "area-job", Region: &logistics.Region{ID: "area-test", Bounds: b}}, nil
}

func (q *queries) GetJob(context.Context, string) (logistics_service.Job, error) {
	return logistics_service.Job{}, logistics_service.ErrNotFound
}

func (q *queries) GetGraph(_ context.Context, id string) (logistics.LogisticsGraph, error) {

	if id == "missing" {
		return logistics.LogisticsGraph{}, logistics_service.ErrNotFound
	}

	return logistics.LogisticsGraph{ID: id, Nodes: []logistics.LogisticsNode{{ID: "n", Modes: []logistics.TransportMode{logistics.TransportRoad}}, {ID: "port", Modes: []logistics.TransportMode{logistics.TransportSea}}}}, nil
}

func TestAPIValidationAndExports(t *testing.T) {

	gin.SetMode(gin.TestMode)
	q := &queries{}
	r := httpapi.NewRouter(q)
	cases := []struct {
		method, path, body string
		status             int
	}{
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31,"north":61}}`, 202},
		{"POST", "/api/graphs", `{"bounds":{"west":31,"south":59,"east":29,"north":61}}`, 400},
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31}}`, 400},
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31,"north":null}}`, 400},
		{"POST", "/api/graphs", `{"region_id":"r","bounds":{"west":29,"south":59,"east":31,"north":61}}`, 400},
		{"POST", "/api/graphs", `{"bounds":{"west":-181,"south":59,"east":31,"north":61}}`, 400},
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

	if q.calls != 2 {
		t.Fatalf("invalid request reached application: %d calls", q.calls)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/graphs/g?mode=road", nil))

	if strings.Contains(w.Body.String(), `"id":"port"`) {
		t.Fatal("mode filter leaked sea node")
	}
}

func TestAreaWithoutSourceReturns422(t *testing.T) {

	gin.SetMode(gin.TestMode)
	router := httpapi.NewRouter(&queries{areaErr: logistics_service.ErrAreaUnavailable})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("POST", "/api/graphs", strings.NewReader(`{"bounds":{"west":1,"south":1,"east":2,"north":2}}`)))

	if response.Code != 422 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
