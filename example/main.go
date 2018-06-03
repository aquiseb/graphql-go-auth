// https://stackoverflow.com/questions/38897529/golang-pass-method-to-function/38897667#38897667
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/skratchdot/open-golang/open"

	googleOAuth2 "golang.org/x/oauth2/google"

	authCommon "github.com/astenmies/graphql-go-auth/authCommon"
	"github.com/astenmies/graphql-go-auth/authGoogle"
	"github.com/astenmies/graphql-go-auth/authUtils"
	"github.com/rs/cors"

	"github.com/gorilla/sessions"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type Config struct {
	ClientID     string
	ClientSecret string
}

var graphqlSchema *graphql.Schema

var customConfig = &authUtils.Config{
	Name:            "graphql-go-auth",
	Path:            "/",
	MaxAge:          60, // 60 seconds
	HTTPOnly:        true,
	Secure:          false,          // allows cookies to be send over HTTP
	TriggerMutation: "triggerOauth", // the mutation that triggers Oauth
}

var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

// Provide a cookie session after successful Google login callback
func callbackSuccess() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		googleUser, err := authGoogle.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a session
		session, _ := sessionStore.New(req, sessionName)
		session.Values[sessionUserKey] = googleUser.Id
		session.Save(req, w)

		// http.Redirect(w, req, "/profile", http.StatusFound)
		// show succes page
		msg := "<p><strong>Hello " + googleUser.GivenName + " " + googleUser.FamilyName + "</strong></p>"
		msg = msg + "<p>You are authenticated!</p>"
		fmt.Fprintf(w, msg)
	}
	return http.HandlerFunc(fn)
}

// When the whole login process is successful, this comes last
func querySuccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		authURL := ctx.Value(authCommon.AuthURLKey).(string)

		// Don't use this in production
		// instead, use http.Redirect and define CORS.
		open.Run(authURL)

		// http.Redirect(w, req, authURL, http.StatusFound)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func main() {

	oauth2Config := &oauth2.Config{
		ClientID:     viper.GetString("gqlauth.oauth.google.id"),
		ClientSecret: viper.GetString("gqlauth.oauth.google.secret"),
		RedirectURL:  "http://localhost:8080/google/callback",
		Endpoint:     googleOAuth2.Endpoint,
		Scopes:       []string{"profile", "email"},
	}

	h := &relay.Handler{Schema: graphqlSchema}
	handleSuccess := querySuccess(h)
	handleLogin := authCommon.LoginHandler(oauth2Config, handleSuccess, nil)
	handleState := authCommon.StateCookieHandler(customConfig, handleLogin, h)
	http.Handle("/graphql", cors.Default().Handler(handleState))

	handleSuccess = callbackSuccess()
	handleGoogle := authGoogle.Handler(oauth2Config, handleSuccess, nil)
	handleCallback := authCommon.CallbackHandler(oauth2Config, handleGoogle, nil)
	handleState = authCommon.StateCookieHandler(customConfig, handleCallback, nil)
	http.Handle("/google/callback", handleState)

	// Write a GraphiQL page to /
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))

	// Start an http server
	log.Println("Let's go")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//// Users ////
var users = []User{
	{
		ID:       "1",
		Username: "bob",
	},
}

//// GraphQL Schema ////

// Schema describes the data that we ask for
var Schema = `
    schema {
		query: Query
        mutation: Mutation
	}
	type Query {
		user(input: UserInput!): String!
	}
	type Mutation {
		triggerOauth(input: UserLoginInput!): String!
	}
	type User {
		id: ID!
		username: String!
		password: String!
	}
	input UserInput {
		username: String!
	}
    input UserLoginInput {
		username: String!
	}
    `

func init() {

	// Define global config
	viper.SetConfigName("global")
	viper.AddConfigPath("_config")

	// Does viper work?
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	// MustParseSchema parses a GraphQL schema and attaches the given root resolver.
	// It returns an error if the Go type signature of the resolvers does not match the schema.
	graphqlSchema = graphql.MustParseSchema(Schema, &Resolver{})

}

//// Resolvers ////
var (
	sessionName    = "gqlauth"
	sessionSecret  = viper.GetString("gqlauth.cookie.secret")
	sessionUserKey = "googleID"
)

// User :
// - Resolves User query
func (r *Resolver) User(ctx context.Context, args *struct {
	Input *User
}) string {
	return "Username"
}

// triggerOauth :
// - Resolves triggerOauth mutation
func (r *Resolver) TriggerOauth(ctx context.Context, args *struct {
	Input *UserLoginInput
}) string {
	return "newToken for " + args.Input.Username
}

//// Graphql Types ////

// Resolver common struct
type Resolver struct{}

// User struct for query
type User struct {
	ID       string
	Username string
}

// UserInput struct for query
type UserInput struct {
	Username string
}

// UserLoginInput struct for userLogin mutation
// while an input is not necessary for the trigger mutation
// it can be convenient to save an oauth token in DB, associated with a username
// but you may also enable to define the username later on (on a profile page for instance)
type UserLoginInput struct {
	Username string
}

//// GraphiQL ////
var page = []byte(`
    <!DOCTYPE html>
    <html>
        <head>
            <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.css" />
            <script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
            <script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
            <script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
            <script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.js"></script>
        </head>
        <body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
            <div id="graphiql" style="height: 100vh;">Loading...</div>
            <script>
                function graphQLFetcher(graphQLParams) {
                    return fetch("/graphql", {
						method: "post",
                        body: JSON.stringify(graphQLParams),
						credentials: "include",
                    }).then(function (response) {
                        return response.text();
                    }).then(function (responseBody) {
                        try {
                            return JSON.parse(responseBody);
                        } catch (error) {
                            return responseBody;
                        }
                    });
                }
                ReactDOM.render(
                    React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
                    document.getElementById("graphiql")
                );
            </script>
        </body>
    </html>
    `)
