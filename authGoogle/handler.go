package authGoogle

import (
	"errors"
	"net/http"

	"github.com/astenmies/graphql-auth/authCommon"
	"github.com/astenmies/graphql-auth/authUtils"
	"golang.org/x/oauth2"
	google "google.golang.org/api/oauth2/v2"
)

var (
	ErrUnableToGetGoogleUser    = errors.New("google: unable to get Google User")
	ErrCannotValidateGoogleUser = errors.New("google: could not validate Google User")
)

// GoogleHandler :
// - Gets the OAuth2 Token from the ctx
// - Then gets Google Userinfoplus with token
// - Adds user info to the ctx and the success handler is called
// - Otherwise, the failure handler is called
func Handler(config *oauth2.Config, success http.Handler, failure http.Handler) http.Handler {
	if failure == nil {
		failure = authUtils.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := authCommon.TokenFromContext(ctx)
		if err != nil {
			ctx = authUtils.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		httpClient := config.Client(ctx, token)
		googleService, err := google.New(httpClient)
		if err != nil {
			ctx = authUtils.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		userInfoPlus, err := googleService.Userinfo.Get().Do()
		err = validateResponse(userInfoPlus, err)
		if err != nil {
			ctx = authUtils.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		ctx = UserToContext(ctx, userInfoPlus)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// validateResponse :
// - Returns an error if no given Google Userinfoplus
// - http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *google.Userinfoplus, err error) error {
	if err != nil {
		return ErrUnableToGetGoogleUser
	}
	if user == nil || user.Id == "" {
		return ErrCannotValidateGoogleUser
	}
	return nil
}
