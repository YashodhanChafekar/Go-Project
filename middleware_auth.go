package main

import (
	"fmt"
	"net/http"

	"github.com/YashodhanChafekar/go_rssagg/internal/auth"
	"github.com/YashodhanChafekar/go_rssagg/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (apiCfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r.Header)
		if err != nil {
			respondWithError(w, http.StatusNetworkAuthenticationRequired, fmt.Sprintf("Auth error: %v", err))
			return
		}

		user, err := apiCfg.DB.GetUserByAPTKey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("User Not Found: %v", err))
			return
		}
		handler(w, r, user)
	}
}
