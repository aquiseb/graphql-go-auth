package authUtils

import (
	"net/http"
	"time"
)

// New :
// - Returns a new http.Cookie with the given value and CookieConfig properties
// - MaxAge to determines if Expires field should be added for Internet Explorer compatability and what its value should be
func NewCookie(config *Config, value string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     config.Name,
		Value:    value,
		Domain:   config.Domain,
		Path:     config.Path,
		MaxAge:   config.MaxAge,
		HttpOnly: config.HTTPOnly,
		Secure:   config.Secure,
	}
	// IE <9 does not understand MaxAge, set Expires if MaxAge is non-zero.
	if expires, ok := expiresTime(config.MaxAge); ok {
		cookie.Expires = expires
	}
	return cookie
}

// expiresTime :
// - Converts a maxAge time in seconds to a time.
// - Time in the future if the maxAge is positive or the beginning of the epoch if maxAge is negative.
// - If maxAge is > 0, time in the future
// - If maxAge is < 0, beginning of the epoch
// - If maxAge is exactly 0, an empty time and false are returned (so the Cookie Expires field should not be set).
func expiresTime(maxAge int) (time.Time, bool) {
	if maxAge > 0 {
		d := time.Duration(maxAge) * time.Second
		return time.Now().Add(d), true
	} else if maxAge < 0 {
		return time.Unix(1, 0), true // first second of the epoch
	}
	return time.Time{}, false
}
