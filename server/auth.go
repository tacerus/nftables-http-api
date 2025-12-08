package server

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/tacerus/nftables-http-api/core"
)

const TOKEN_HEADER = "X-NFT-API-TOKEN"

func doCheckToken(token string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	if err == nil {
		return true
	} else {
		slog.Debug("Token check failed", "error", err)
		return false
	}
}

var (
	freePaths = []string{
		"/",
	}
)

func (app *App) authHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(freePaths, r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		var givenToken = r.Header.Get(TOKEN_HEADER)

		if givenToken == "" {
			app.errorHandler(w, http.StatusUnauthorized, "Missing token.")
			return
		}

		var validToken string
		var validPaths core.ConfigTokenPaths
		for configToken, configPaths := range app.tokens {
			slog.Debug("testing token", "request", givenToken, "config", configToken)
			if doCheckToken(givenToken, configToken) {
				validToken = configToken
				validPaths = configPaths
				break
			}
		}

		if len(validPaths) == 0 {
			if validToken == "" {
				app.errorHandler(w, http.StatusUnauthorized, "Invalid token.")
			} else {
				app.errorHandler(w, http.StatusUnauthorized, "The supplied token is not fully configured.")
			}
			return
		}

		var requestResource string
		requestPathElements := strings.Split(r.URL.Path, "/")
		requestResource, requestPathElements = requestPathElements[0], requestPathElements[1:]

	paths:
		for configPath, configMethods := range validPaths {
			configPathElements := strings.Split(configPath, "/")
			var configResource string
			configResource, configPathElements = configPathElements[0], configPathElements[1:]

			// short-circuit
			if configResource != requestResource {
				continue
			}

			for i := range len(configPathElements) {
				if configPathElements[i] == "*" {
					continue
				}
				if configPathElements[i] == requestPathElements[i] {
					continue
				}

				continue paths
			}

			for _, method := range configMethods {
				if method == r.Method {
					handler.ServeHTTP(w, r)
					return
				}
			}

			continue paths
		}

		slog.Debug("Token not authorized", "source", r.RemoteAddr, "method", r.Method, "path", r.URL.Path, "token", validToken)
		app.errorHandler(w, http.StatusUnauthorized, "The supplied token is not authorized to perform this request.")
		return

	})
}
