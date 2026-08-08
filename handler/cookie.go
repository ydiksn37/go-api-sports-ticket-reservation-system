package handler

import (
	"42tokyo-road-to-dena-server/authbundle"
	"net/http"
)

// アクセストークンとリフレッシュトークンを Cookie にセットする
func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, cfg *authbundle.AuthConfig) {
	// アクセストークン Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(cfg.AccessTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// リフレッシュトークン Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(cfg.RefreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
