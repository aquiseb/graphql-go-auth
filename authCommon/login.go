package authCommon

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	authUtils "github.com/astenmies/graphql-go-auth/authUtils"
	graphql "github.com/graph-gophers/graphql-go"
	"golang.org/x/oauth2"
)

// Returns a base64 encoded random 32 byte string
func randomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Error messages
var (
	ErrInvalidState = errors.New("oauth2: Invalid OAuth2 state parameter")
)

type testr struct {
	Query string
}

type Handler struct {
	Schema *graphql.Schema
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func isTrigger(config *authUtils.Config, query string) bool {
	// If our string contains "IntrospectionQuery"
	// it means we're not analysing the right graphql query
	if !strings.Contains(query, "IntrospectionQuery") {
		// Replace all special characters
		var replacer = strings.NewReplacer(" ", "", "\n", "", "(", "|", "{", "|") // each odd is the char to be replaced
		var result = replacer.Replace(query)

		// Split the new string
		var resultSlice = strings.Split(result, "|")
		var isMutation = contains(resultSlice, "mutation")
		var mutationName string

		// Does our query contain mutation?
		if isMutation {
			mutationName = resultSlice[1]
			if mutationName == config.TriggerMutation {
				return true
			}
		}
	}
	return false
}

// StateCookieHandler :
// - Oauth2 requires a state
// - If state cookie exists, read its value and add to ctx var
// - Otherwise generate a random value and add it to ctx
// - Takes four args:
//		1- your auth config
//		2- success is the function that is called after successful state management
//		3- normalQuery bypasses the success function if it should not get called (means it's not the mutation that triggers oauth)
func StateCookieHandler(config *authUtils.Config, success http.Handler, normalQuery http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		cookie, err := req.Cookie(config.Name)

		if err == nil { // No error means we have a cookie -> add it to our state
			ctx = StateToContext(ctx, cookie.Value)
		} else {
			// Generate a random state and store it in cookie
			val := randomState()
			http.SetCookie(w, authUtils.NewCookie(config, val))
			ctx = StateToContext(ctx, val)
		}

		// Now here below, let's see what we should return
		// If we're dealing with the triggerMutation, then we should continue with success handler
		// Otherwise we should continue with the relay handler
		// Read the content
		buf, _ := ioutil.ReadAll(req.Body)

		rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
		rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
		// Restore the to its original state
		req.Body = rdr2
		// manipulate rd1 only
		// log.Printf("BODY: %q", rdr1)
		decoder := json.NewDecoder(rdr1)

		var t struct {
			Query string `json:"query"`
		}
		decoder.Decode(&t)

		// Todo replace string by a value that should be defined in the config
		// If our mutation name is the one that should trigger,
		// we continue the HandlerFunc chain success.serveHTTP
		if isTrigger(config, t.Query) || normalQuery == nil {
			success.ServeHTTP(w, req.WithContext(ctx))
		} else {
			normalQuery.ServeHTTP(w, req.WithContext(ctx))
		}

	}
	return http.HandlerFunc(fn)
}

// LoginHandler :
// - Reads the state value from ctx
// - Executes success function if passed
// - Otherwise redirects requests to the AuthURL with the state value.
func LoginHandler(config *oauth2.Config, success http.Handler, failure http.Handler) http.Handler {
	if failure == nil {
		failure = authUtils.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		// Extract the state from our ctx
		state, err := StateFromContext(ctx)

		if err != nil {
			ctx = authUtils.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		authURL := config.AuthCodeURL(state)
		ctx = AuthURLToContext(ctx, authURL)

		// If no success handler is passed, use the default redirection
		if success == nil {
			http.Redirect(w, req, authURL, http.StatusFound)
		} else {
			success.ServeHTTP(w, req.WithContext(ctx))
		}

	}
	return http.HandlerFunc(fn)
}

// CallbackHandler :
// - Checks for a state cookie
// - Adds state value to ctx
func CallbackHandler(config *oauth2.Config, success http.Handler, failure http.Handler) http.Handler {
	if failure == nil {
		failure = authUtils.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		authCode, state, err := StateAndCodeFromReq(req)
		if err != nil {
			ctx = authUtils.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ownerState, err := StateFromContext(ctx)
		if err != nil {
			ctx = authUtils.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		if state != ownerState || state == "" {
			ctx = authUtils.WithError(ctx, ErrInvalidState)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		// Ask for a token with the authorization code
		token, err := config.Exchange(ctx, authCode)
		if err != nil {
			ctx = authUtils.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		ctx = TokenToContext(ctx, token)
		success.ServeHTTP(w, req.WithContext(ctx))

	}
	return http.HandlerFunc(fn)
}
