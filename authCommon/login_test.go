package authCommon

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/astenmies/graphql-go-auth/authUtils"
	"github.com/stretchr/testify/assert"
)

func AssertSuccess(t *testing.T) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Success handler called")
	}
	return http.HandlerFunc(fn)
}

func Test_StateCookieHandler(t *testing.T) {
	config := &authUtils.Config{
		Name:     "gqlauth_cookie",
		Domain:   "dom",
		Path:     "/",
		MaxAge:   100,
		HTTPOnly: false,
		Secure:   false,
	}

	success := AssertSuccess(t)
	StateCookieHandler := StateCookieHandler(config, success)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	StateCookieHandler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Success handler called", w.Body.String())
}
