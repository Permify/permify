package integration_test

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
)

const (
	// Attempts connection
	host       = "permify:3476"
	healthPath = "http://" + host + "/healthz"
	attempts   = 20

	// HTTP REST
	basePath = "http://" + host + "/v1"
)

func TestMain(m *testing.M) {
	err := healthCheck(attempts)
	if err != nil {
		log.Fatalf("Integration tests: host %s is not available: %s", host, err)
	}

	log.Printf("Integration tests: host %s is available", host)

	code := m.Run()
	os.Exit(code)
}

func healthCheck(attempts int) error {
	var err error

	for attempts > 0 {
		err = Do(Get(healthPath), Expect().Status().Equal(http.StatusOK))
		if err == nil {
			return nil
		}

		log.Printf("Integration tests: url %s is not available, attempts left: %d", healthPath, attempts)

		time.Sleep(time.Second)

		attempts--
	}

	return err
}

// HTTP POST: /translation/do-translate.
func TestHTTPCheckRequest(t *testing.T) {

	body := `{
    	"schema": "entity user {} \n\nentity account {\n    // roles \n    relation admin @user    \n    relation member @user    \n    relation parent_account @account\n\n    action add_member = admin or parent_account.add_member\n    action delete_member = admin\n\n}"
	}`

	Test(t,
		Description("Create Schema"),
		Post(basePath+"/tenants/t1/schemas/write"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
	)

	body = `{
    	"metadata": {
        	"snap_token": "",
        	"schema_version": "",
        	"depth": 100
    	},
    	"entity": {
        	"type": "account",
        	"id": "r1"
    	},
    	"permission": "add_member",
    	"subject": {
        	"type": "user",
        	"id": "u1"
    	}
	}`
	Test(t,
		Description("Check"),
		Post(basePath+"/tenants/t1/permissions/check"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".can").Equal("RESULT_DENIED"),
	)

}
