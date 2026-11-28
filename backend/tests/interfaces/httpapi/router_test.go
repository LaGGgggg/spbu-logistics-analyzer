package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/account"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/logistics"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/interfaces/httpapi"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/logistics_service"
	"github.com/gin-gonic/gin"
)

type queries struct {
	calls       int
	cancelCalls int
	areaErr     error
	cancelErr   error
}

func (q *queries) Regions(context.Context) ([]logistics.Region, error) {
	return []logistics.Region{}, nil
}

func (q *queries) RequestBuild(context.Context, string) (logistics_service.BuildResult, error) {

	q.calls++

	return logistics_service.BuildResult{JobID: "job"}, nil
}

func (q *queries) RequestAreaBuild(_ context.Context, b logistics.Bounds, _ *int) (logistics_service.BuildResult, error) {

	q.calls++

	if q.areaErr != nil {
		return logistics_service.BuildResult{}, q.areaErr
	}

	return logistics_service.BuildResult{JobID: "area-job", Region: &logistics.Region{ID: "area-test", Bounds: b}}, nil
}

func (q *queries) GetJob(context.Context, string) (logistics_service.Job, error) {
	return logistics_service.Job{}, logistics_service.ErrNotFound
}

func (q *queries) CancelJob(_ context.Context, id string) (logistics_service.Job, error) {

	q.cancelCalls++

	if q.cancelErr != nil {
		return logistics_service.Job{}, q.cancelErr
	}

	return logistics_service.Job{ID: id, RegionID: "r", Status: logistics_service.Cancelled}, nil
}

func (q *queries) GetGraph(_ context.Context, id string) (logistics.LogisticsGraph, error) {

	if id == "missing" {
		return logistics.LogisticsGraph{}, logistics_service.ErrNotFound
	}

	return logistics.LogisticsGraph{ID: id, Nodes: []logistics.LogisticsNode{{ID: "n", Modes: []logistics.TransportMode{logistics.TransportRail}}, {ID: "port", Modes: []logistics.TransportMode{logistics.TransportSea}}}}, nil
}

type testAuth struct{}

func (testAuth) Register(_ context.Context, _ account.Credentials, _ string) (account.User, account.Session, error) {
	return account.User{ID: "user_1", Email: "user@example.com", Name: "user"}, account.Session{ID: "session-1"}, nil
}

func (testAuth) Login(_ context.Context, _ account.Credentials, _ string) (account.User, account.Session, error) {
	return account.User{ID: "user_1", Email: "user@example.com", Name: "user"}, account.Session{ID: "session-1"}, nil
}

func (testAuth) Current(_ context.Context, sessionID string) (account.User, error) {

	if sessionID != "session-1" {
		return account.User{}, account.ErrUnauthorized
	}

	return account.User{ID: "user_1", Email: "user@example.com", Name: "user"}, nil
}

func (testAuth) Logout(context.Context, string) error {
	return nil
}

func TestAPIValidationAndExports(t *testing.T) {

	gin.SetMode(gin.TestMode)
	q := &queries{}
	r := httpapi.NewRouter(q, testAuth{}, httpapi.AuthOptions{})
	cases := []struct {
		method, path, body string
		status             int
	}{
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31,"north":61}}`, 202},
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31,"north":61},"optimization":0}`, 202},
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31,"north":61},"optimization":101}`, 400},
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31,"north":61},"optimization":-1}`, 400},
		{"POST", "/api/graphs", `{"bounds":{"west":31,"south":59,"east":29,"north":61}}`, 400},
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31}}`, 400},
		{"POST", "/api/graphs", `{"bounds":{"west":29,"south":59,"east":31,"north":null}}`, 400},
		{"POST", "/api/graphs", `{"region_id":"r","bounds":{"west":29,"south":59,"east":31,"north":61}}`, 400},
		{"POST", "/api/graphs", `{"bounds":{"west":-181,"south":59,"east":31,"north":61}}`, 400},
		{"GET", "/health", "", 200}, {"POST", "/api/graphs", `{"region_id":"r"}`, 202},
		{"POST", "/api/graphs", `{"region_id":"r","url":"http://example.com"}`, 400},
		{"POST", "/api/graphs", `{"region_id":"r"} {}`, 400}, {"POST", "/api/graphs", `{"region_id":" r"}`, 400},
		{"POST", "/api/graphs", `null`, 400}, {"POST", "/api/graphs", strings.Repeat(" ", 5000), 400},
		{"GET", "/api/jobs/missing", "", 404}, {"POST", "/api/jobs/job/cancel", "", 200}, {"GET", "/api/graphs/missing", "", 404},
		{"GET", "/api/graphs/g?mode=invalid", "", 400}, {"GET", "/api/graphs/g/export/geojson", "", 200},
		{"GET", "/api/graphs/g/export/bad", "", 400}, {"GET", "/api/graphs/g/coverage", "", 200},
		{"GET", "/api/graphs/g/metrics", "", 200}, {"GET", "/api/graphs/missing/metrics", "", 404},
		{"GET", "/api/graphs/g/metrics?mode=rail", "", 200}, {"GET", "/api/graphs/g/metrics?mode=invalid", "", 400},
	}

	for _, c := range cases {
		w := httptest.NewRecorder()
		request := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
		request.AddCookie(&http.Cookie{Name: "magistral_session", Value: "session-1"})
		r.ServeHTTP(w, request)

		if w.Code != c.status {
			t.Errorf("%s %s => %d: %s", c.method, c.path, w.Code, w.Body.String())
		}
	}

	if q.calls != 3 {
		t.Fatalf("invalid request reached application: %d calls", q.calls)
	}

	if q.cancelCalls != 1 {
		t.Fatalf("cancel requests=%d", q.cancelCalls)
	}

	for _, path := range []string{"/api/graphs"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(`{"region_id":"r"}`)))

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("anonymous %s => %d: %s", path, w.Code, w.Body.String())
		}
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/api/jobs/job/cancel", nil))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous cancel => %d: %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/graphs/g?mode=rail", nil))

	if strings.Contains(w.Body.String(), `"id":"port"`) {
		t.Fatal("mode filter leaked sea node")
	}
}

func TestCancelCompletedJobReturns409(t *testing.T) {

	gin.SetMode(gin.TestMode)
	router := httpapi.NewRouter(&queries{cancelErr: logistics_service.ErrJobNotCancellable}, testAuth{}, httpapi.AuthOptions{})
	request := httptest.NewRequest("POST", "/api/jobs/done/cancel", nil)
	request.AddCookie(&http.Cookie{Name: "magistral_session", Value: "session-1"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != 409 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAreaWithoutSourceReturns422(t *testing.T) {

	gin.SetMode(gin.TestMode)
	router := httpapi.NewRouter(&queries{areaErr: logistics_service.ErrAreaUnavailable}, testAuth{}, httpapi.AuthOptions{})
	request := httptest.NewRequest("POST", "/api/graphs", strings.NewReader(`{"bounds":{"west":1,"south":1,"east":2,"north":2}}`))
	request.AddCookie(&http.Cookie{Name: "magistral_session", Value: "session-1"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != 422 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

type failingAuth struct {
	testAuth
	err error
}

func (f failingAuth) Login(_ context.Context, _ account.Credentials, _ string) (account.User, account.Session, error) {
	return account.User{}, account.Session{}, f.err
}

func TestAuthEndpoints(t *testing.T) {

	gin.SetMode(gin.TestMode)
	router := httpapi.NewRouter(&queries{}, testAuth{}, httpapi.AuthOptions{})
	post := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(body)))

		return w
	}

	w := post("/api/auth/login", `{"email":"user@example.com","password":"password123"}`)

	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"email":"user@example.com"`) {
		t.Fatalf("login status=%d body=%s", w.Code, w.Body.String())
	}

	var session *http.Cookie

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "magistral_session" {
			session = cookie
		}
	}

	if session == nil || session.Value != "session-1" {
		t.Fatalf("session cookie not issued: %+v", w.Result().Cookies())
	}

	if !session.HttpOnly || session.SameSite != http.SameSiteLaxMode || session.Secure || session.Path != "/" {
		t.Fatalf("insecure session cookie: %+v", session)
	}

	if strings.Contains(w.Body.String(), "password") {
		t.Fatalf("credentials leaked in response: %s", w.Body.String())
	}

	w = post("/api/auth/register", `{"email":"new@example.com","password":"password123"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", w.Code, w.Body.String())
	}

	for _, c := range []struct {
		path, body string
		status     int
	}{
		{"/api/auth/login", `{"email":"user@example.com"}`, http.StatusBadRequest},
		{"/api/auth/login", `{"password":"password123"}`, http.StatusBadRequest},
		{"/api/auth/login", `{"email":"a@b.c","password":"x","admin":true}`, http.StatusBadRequest},
		{"/api/auth/login", `not json`, http.StatusBadRequest},
	} {
		if w := post(c.path, c.body); w.Code != c.status {
			t.Errorf("%s %s => %d, want %d: %s", c.path, c.body, w.Code, c.status, w.Body.String())
		}
	}

	me := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "magistral_session", Value: "session-1"})
	router.ServeHTTP(me, request)

	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), `"email":"user@example.com"`) {
		t.Fatalf("me status=%d body=%s", me.Code, me.Body.String())
	}

	me = httptest.NewRecorder()
	router.ServeHTTP(me, httptest.NewRequest("GET", "/api/auth/me", nil))

	if me.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous me status=%d", me.Code)
	}

	logout := httptest.NewRecorder()
	request = httptest.NewRequest("POST", "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: "magistral_session", Value: "session-1"})
	router.ServeHTTP(logout, request)

	cleared := false

	for _, cookie := range logout.Result().Cookies() {
		if cookie.Name == "magistral_session" && cookie.Value == "" && cookie.MaxAge < 0 {
			cleared = true
		}
	}

	if logout.Code != http.StatusOK || !cleared {
		t.Fatalf("logout status=%d cookies=%+v", logout.Code, logout.Result().Cookies())
	}
}

func TestAuthErrors(t *testing.T) {

	gin.SetMode(gin.TestMode)

	for _, c := range []struct {
		err    error
		status int
	}{
		{account.ErrInvalidCredentials, http.StatusUnauthorized},
		{account.ErrInvalidInput, http.StatusBadRequest},
		{account.ErrEmailTaken, http.StatusConflict},
		{errors.New("database down"), http.StatusInternalServerError},
	} {
		router := httpapi.NewRouter(&queries{}, failingAuth{err: c.err}, httpapi.AuthOptions{})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"a@b.c","password":"password123"}`)))

		if w.Code != c.status {
			t.Errorf("%v => %d, want %d: %s", c.err, w.Code, c.status, w.Body.String())
		}

		if strings.Contains(w.Body.String(), "database down") {
			t.Errorf("internal error leaked: %s", w.Body.String())
		}
	}
}

func TestAuthSecureCookieOption(t *testing.T) {

	gin.SetMode(gin.TestMode)
	router := httpapi.NewRouter(&queries{}, testAuth{}, httpapi.AuthOptions{SecureCookies: true})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"a@b.c","password":"password123"}`)))

	secure := false

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "magistral_session" && cookie.Secure {
			secure = true
		}
	}

	if !secure {
		t.Fatalf("SecureCookies not applied: %+v", w.Result().Cookies())
	}
}
