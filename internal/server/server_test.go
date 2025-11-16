package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/iperez/new-expenses-go/internal/config"
	"github.com/iperez/new-expenses-go/internal/server"
)

func TestExpensesFlow(t *testing.T) {
	db := newTestDB(t)
	redisClient := newTestRedis(t)

	srv := server.New(testConfig(), db, redisClient)
	app := srv.App()

	user := createUser(t, app)

	sessionCookie := login(t, app, user.Email, "secret123")
	category := createCategory(t, app, sessionCookie)
	createTransaction(t, app, sessionCookie, category.CategoryID)

	list := listTransactions(t, app, sessionCookie)
	require.Len(t, list, 1)
	require.NotNil(t, list[0].CategoryID)
	require.Equal(t, category.CategoryID, *list[0].CategoryID)
}

type userPayload struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type categoryPayload struct {
	CategoryID string `json:"categoryId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}

type transactionPayload struct {
	TransactionID string  `json:"transactionId"`
	CategoryID    *string `json:"categoryId"`
}

type customResponse struct {
	Result bool            `json:"result"`
	Data   json.RawMessage `json:"data"`
}

func createUser(t *testing.T, app *fiber.App) userPayload {
	body := map[string]string{
		"email":     "demo@example.com",
		"firstName": "Demo",
		"lastName":  "User",
		"password":  "secret123",
	}

	resp := doRequest(t, app, http.MethodPost, "/api/users", body, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var parsed customResponse
	decodeResponse(t, resp.Body, &parsed)

	var user userPayload
	require.NoError(t, json.Unmarshal(parsed.Data, &user))

	return user
}

func login(t *testing.T, app *fiber.App, email, password string) *http.Cookie {
	body := map[string]string{
		"email":    email,
		"password": password,
	}

	resp := doRequest(t, app, http.MethodPost, "/api/auth/login", body, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	cookie := findCookie(resp.Cookies(), "sessionID")
	require.NotNil(t, cookie)

	return cookie
}

func createCategory(t *testing.T, app *fiber.App, cookie *http.Cookie) categoryPayload {
	body := map[string]string{
		"name": "Salary",
		"type": "INCOME",
	}

	resp := doRequest(t, app, http.MethodPost, "/api/categories", body, []*http.Cookie{cookie})
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var parsed customResponse
	decodeResponse(t, resp.Body, &parsed)

	var category categoryPayload
	require.NoError(t, json.Unmarshal(parsed.Data, &category))

	return category
}

func createTransaction(t *testing.T, app *fiber.App, cookie *http.Cookie, categoryID string) {
	ex := 40.0
	body := map[string]interface{}{
		"type":         "INCOME",
		"amount":       1000,
		"currency":     "USD",
		"note":         "Salary",
		"day":          1,
		"month":        "JANUARY",
		"year":         2024,
		"exchangeRate": ex,
		"categoryId":   categoryID,
	}

	resp := doRequest(t, app, http.MethodPost, "/api/transactions", body, []*http.Cookie{cookie})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func listTransactions(t *testing.T, app *fiber.App, cookie *http.Cookie) []transactionPayload {
	resp := doRequest(t, app, http.MethodGet, "/api/transactions", nil, []*http.Cookie{cookie})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed customResponse
	decodeResponse(t, resp.Body, &parsed)

	var txs []transactionPayload
	require.NoError(t, json.Unmarshal(parsed.Data, &txs))

	return txs
}

func doRequest(t *testing.T, app *fiber.App, method, path string, body interface{}, cookies []*http.Cookie) *http.Response {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	return resp
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}

	return nil
}

func decodeResponse(t *testing.T, body io.ReadCloser, out interface{}) {
	defer body.Close()
	decoder := json.NewDecoder(body)
	require.NoError(t, decoder.Decode(out))
}

func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)

	return db
}

func newTestRedis(t *testing.T) *redis.Client {
	mr := miniredis.RunT(t)

	opts, err := redis.ParseURL(fmt.Sprintf("redis://%s", mr.Addr()))
	require.NoError(t, err)

	return redis.NewClient(opts)
}

func testConfig() config.Config {
	return config.Config{
		Env:               "TEST",
		Port:              0,
		DatabaseURL:       "",
		RedisURL:          "",
		SessionCookieName: "sessionID",
		SessionTTL:        24 * time.Hour,
		CorsOrigins:       []string{"http://test"},
	}
}
