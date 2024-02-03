// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
    "mooon-gateway/pb/mooon_auth"
    "net/http"
    "time"
)
import (
    "mooon-gateway/pb/mooon_login"
)

func LoginCookie2HttpCookie(loginCookie *mooon_login.Cookie) *http.Cookie {
    httpCookie := &http.Cookie{
        Name:     loginCookie.Name,
        Value:    loginCookie.Value,
        Path:     loginCookie.Path,
        Domain:   loginCookie.Domain,
        Expires:  time.Unix(loginCookie.Expires, 0),
        MaxAge:   int(loginCookie.MaxAge),
        Secure:   loginCookie.Secure,
        HttpOnly: loginCookie.HttpOnly,
    }
    return httpCookie
}

func AuthCookie2HttpCookie(authCookie *mooon_auth.Cookie) *http.Cookie {
    httpCookie := &http.Cookie{
        Name:     authCookie.Name,
        Value:    authCookie.Value,
        Path:     authCookie.Path,
        Domain:   authCookie.Domain,
        Expires:  time.Unix(authCookie.Expires, 0),
        MaxAge:   int(authCookie.MaxAge),
        Secure:   authCookie.Secure,
        HttpOnly: authCookie.HttpOnly,
    }
    return httpCookie
}

func HttpCookie2AuthCookie(httpCookie *http.Cookie) *mooon_auth.Cookie {
    return &mooon_auth.Cookie{
        Name:     httpCookie.Name,
        Value:    httpCookie.Value,
        Path:     httpCookie.Path,
        Domain:   httpCookie.Domain,
        Expires:  httpCookie.Expires.Unix(),
        MaxAge:   int32(httpCookie.MaxAge),
        Secure:   httpCookie.Secure,
        HttpOnly: httpCookie.HttpOnly,
    }
}
