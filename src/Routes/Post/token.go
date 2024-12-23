package post

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	helpers "goserver/src/Helpers"

	http "github.com/bogdanfinn/fhttp"
	"github.com/golang-jwt/jwt/v4"
)

func TokenHandler(logger *helpers.ColorizedLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("Received request with method: %s", r.Method))
		if r.Method != http.MethodPost {
			logger.Error("Invalid method posted to /token")
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("Failed to parse request body")
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		logger.Info("Received Request [Bearer Token Generator]")

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to marshal payload: %s", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		sha256Hash := fmt.Sprintf("%x", sha256.Sum256(payloadBytes))
		secretKey := helpers.GetSecretKey()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"hash": sha256Hash,
			"exp":  time.Now().Add(24 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString([]byte(secretKey))
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to sign JWT token: %s", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		logger.Http("Authorization Bearer Token Generated")
		response := map[string]interface{}{
			"token":   tokenString,
			"message": "Successful Validation",
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error(fmt.Sprintf("Failed to write response: %s", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
