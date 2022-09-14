// main_test.go

package main_test

import (
	"bytes"
	"encoding/json"
	"github.com/cliveyg/poptape-categories"
	"github.com/jarcoal/httpmock"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a main.App

func TestMain(m *testing.M) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	a = main.App{}
	a.Initialize(
		os.Getenv("DB_HOST"),
		os.Getenv("TESTDB_USERNAME"),
		os.Getenv("TESTDB_PASSWORD"),
		os.Getenv("TESTDB_NAME"))

	//ensureTableExists()
	runSQL(tableCreationQuery)

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM categories")
}

func runSQL(sqltext string) {
	if _, err := a.DB.Exec(sqltext); err != nil {
		log.Fatal(err)
	}
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS categories
(
    review_id CHAR(36) UNIQUE NOT NULL,
    public_id CHAR(36) NOT NULL,
    auction_id CHAR(36) NOT NULL,
    review VARCHAR(2000),
    overall INT NOT NULL DEFAULT 0,
    pap_cost INT NOT NULL DEFAULT 0,
    communication INT NOT NULL DEFAULT 0,
    as_described INT NOT NULL DEFAULT 0,
    created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (public_id, auction_id),
    CONSTRAINT reviews_pkey PRIMARY KEY (review_id)
)`

const insertDummyReviews = `INSERT INTO reviews 
(review_id, public_id, auction_id, review, overall, pap_cost, communication, as_described)
VALUES
('e8f48256-2460-418f-81b7-86dad2aa6e41',
'f38ba39a-3682-4803-a498-659f0bf05000',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c',
'amaze balls product',5,4,4,3),
('e8f48256-2460-418f-81b7-86dad2aa6aaa',
'f38ba39a-3682-4803-a498-659f0bf05304',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c',
'amaze balls product',5,4,4,3),
('e8f48256-2460-418f-81b7-86dad2aa6111',
'f38ba39a-3682-4803-a498-659f0bf05304',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab11',
'amaze balls product',4,4,4,3),
('e8f48256-2460-418f-81b7-86dad2aa6222',
'f38ba39a-3682-4803-a498-659f0bf05304',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab22',
'amaze balls product',4,4,4,3),
('e8f48256-2460-418f-81b7-86dad2aa6333',
'f38ba39a-3682-4803-a498-659f0bf05000',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab33',
'amaze balls product',4,4,4,3);`

const createJson = `{"auction_id":"f38ba39a-3682-4803-a498-659f0b111111",
"review": "amazing product",
"overall": 4,
"post_and_packaging": 3,
"communication": 4,
"as_described": 4}`

type Review struct {
	ReviewId  string `json:"review_id"`
	Review    string `json:"review"`
	PublicId  string `json:"public_id"`
	AuctionId string `json:"auction_id"`
	Overall   int    `json:"overall"`
	PapCost   int    `json:"post_and_packaging"`
	Comm      int    `json:"communication"`
	AsDesc    int    `json:"as_described"`
	Created   string `json:"created"`
}

type CreateResp struct {
	ReviewId string `json:"review_id"`
}

// ----------------------------------------------------------------------------
// s t a r t   o f   t e s t s
// ----------------------------------------------------------------------------

func TestAPIStatus(t *testing.T) {

	req, _ := http.NewRequest("GET", "/reviews/status", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

}

// get no reviews for authed user
func TestEmptyTable(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

// get reviews for authed user
func TestReturnOnlyAuthUserReviews(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	reviews := make([]Review, 0)
	json.NewDecoder(response.Body).Decode(&reviews)
	for _, r := range reviews {
		if r.PublicId != "f38ba39a-3682-4803-a498-659f0bf05304" {
			t.Errorf("public id doesn't match")
		}
	}

	if len(reviews) != 3 {
		t.Errorf("no of reviews returned doesn't match")
	}

}

// test missing access token
func TestMissingXAccessToken(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

// get reviews by user - no auth needed
func TestGetReviewsByUser(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	reviews := make([]Review, 0)
	json.NewDecoder(response.Body).Decode(&reviews)
	for _, r := range reviews {
		if r.PublicId != "f38ba39a-3682-4803-a498-659f0bf05304" {
			t.Errorf("public id doesn't match")
		}
	}

	if len(reviews) != 3 {
		t.Errorf("no of reviews returned doesn't match")
	}

}

// test bad uuid
func TestBadUUID(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/user/f38ba39a-3682-4803-a498-659f0bf0530g", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

}

// test 404
func Test404ForValidUUID(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/f38ba39a-3682-4803-a498-659f0bf05311", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

}

// test 404
func Test404ForRandomURL(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/f38ba39a/someurl/999", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

}

// get reviews by auction - no auth needed
func TestGetReviewsByAuction(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/auction/e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	reviews := make([]Review, 0)
	json.NewDecoder(response.Body).Decode(&reviews)

	if len(reviews) != 2 {
		t.Errorf("no of reviews returned doesn't match")
	}

}

// get review by id - no auth needed
func TestGetReviewsById(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6333", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	var rev Review
	json.NewDecoder(response.Body).Decode(&rev)

	if rev.PublicId != "f38ba39a-3682-4803-a498-659f0bf05000" {
		t.Errorf("public id doesn't match")
	}

}

// get delete review for authed user
func TestDeleteReviewOk(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusGone, response.Code)

	req, _ = http.NewRequest("GET", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response = executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

}

// failed delete review - cannot delete someone elses review
func TestDeleteFail(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6333", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotAcceptable, response.Code)

}

// failed delete review when unauthorised
func TestDeleteNotAuthedFail(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(401, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)

}

// test review creation
func TestCreateReviewOk(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))
	url := "https://poptape.club/auctionhouse/auction/f38ba39a-3682-4803-a498-659f0b111111"
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, `{"message": "whatevs"}`))

	//auction_id, review, overall, pap_cost, communication, as_described)
	payload := []byte(createJson)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)
	var crep CreateResp
	json.NewDecoder(response.Body).Decode(&crep)

	req, _ = http.NewRequest("GET", "/reviews/"+crep.ReviewId, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var rev Review
	json.NewDecoder(response.Body).Decode(&rev)
	if rev.PublicId != "f38ba39a-3682-4803-a498-659f0bf05304" {
		t.Errorf("public id doesn't match")
	}
	if rev.AuctionId != "f38ba39a-3682-4803-a498-659f0b111111" {
		t.Errorf("auction id doesn't match")
	}
	if rev.Review != "amazing product" {
		t.Errorf("review doesn't match")
	}
	if rev.Overall != 4 {
		t.Errorf("overall score doesn't match")
	}
	if rev.PapCost != 3 {
		t.Errorf("p&p score doesn't match")
	}
	if rev.Comm != 4 {
		t.Errorf("comm score doesn't match")
	}
	if rev.AsDesc != 4 {
		t.Errorf("as described score doesn't match")
	}
}

// test review creation fails if duplicate attempted
func TestCreateReviewDuplicateReviewFail(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))
	url := "https://poptape.club/auctionhouse/auction/f38ba39a-3682-4803-a498-659f0b111111"
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, `{"message": "whatevs"}`))

	payload := []byte(createJson)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response = executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

}

// test review creation fails if 'overall' field is not numeric
func TestCreateReviewFailOnOverall(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))
	url := "https://poptape.club/auctionhouse/auction/f38ba39a-3682-4803-a498-659f0b111111"
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, `{"message": "whatevs"}`))

	var badOverall = `{"auction_id":"f38ba39a-3682-4803-a498-659f0b111111",
"review": "amazing product",
"overall": "a",
"post_and_packaging": 3,
"communication": 4,
"as_described": 4}`

	payload := []byte(badOverall)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}
