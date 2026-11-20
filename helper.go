package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/OliverMackie/chirpy/internal/auth"
	"github.com/google/uuid"

	"github.com/OliverMackie/chirpy/internal/database"
)

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type APIConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
	plat           string
	SecretKey      string
	APIkey         string
}

func (c *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
func (c *APIConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf("<html>\n	<body>\n		<h1>Welcome, Chirpy Admin</h1>\n		<p>Chirpy has been visited %d times!</p>\n	</body>\n</html>", c.fileserverHits.Load())
	w.Write([]byte(response))
}

func (c *APIConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	if c.plat != "dev" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
	} else {

		err := c.database.Reset(r.Context())
		if err != nil {
			log.Printf("failed to reset database %v", err)
		}
		c.fileserverHits.Store(0)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metrics reset"))
	}

}

func (c *APIConfig) handlerValidate(w http.ResponseWriter, r *http.Request) {
	authHeader, err := auth.GetBearerToken(r.Header)

	if (authHeader == "") || (err != nil) {
		w.WriteHeader(401)
		w.Write([]byte("Missing or invalid Authorization header"))
		return
	}
	token, err := auth.ValidateJWT(authHeader, c.SecretKey)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Invalid token"))
		return
	}
	type body struct {
		Body    string    `json:"body"`
		User_id uuid.UUID `json:"user_id"`
	}
	var b body
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&b)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
	}
	if len(b.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp body exceeds 140 characters")
	} else {
		respondWithJSON(w, http.StatusOK, cleanString(b.Body), token, c, r)
	}

}

func (c *APIConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type email struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var e email
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&e)
	if err != nil {
		log.Println("Invalid request body")
	}
	pass, err := auth.HashPassword(e.Password)
	if err != nil {
		log.Println("Failed hashing password")
	}
	user, err := c.database.CreateUser(r.Context(), database.CreateUserParams{
		Email:          e.Email,
		HashedPassword: pass,
	})
	if err != nil {
		log.Println("Failed creating user")
	}
	type successResponse struct {
		Id            uuid.UUID `json:"id"`
		Created_at    time.Time `json:"created_at"`
		Updated_at    time.Time `json:"updated_at"`
		Email         string    `json:"email"`
		Is_chirpy_red bool      `json:"is_chirpy_red"`
	}
	response := successResponse{
		Id: user.ID, Created_at: user.CreatedAt, Updated_at: user.UpdatedAt, Email: user.Email, Is_chirpy_red: user.IsChirpyRed,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (c *APIConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps := Setchirps(c, r)
	s := r.URL.Query().Get("sort")
	if s == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}
	type successResponse struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Body       string    `json:"body"`
		User_id    uuid.UUID `json:"user_id"`
	}

	returnSlice := make([]successResponse, 0)
	for _, chirp := range chirps {
		response := successResponse{
			Id:         chirp.ID,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			Body:       chirp.Body,
			User_id:    chirp.UserID,
		}
		returnSlice = append(returnSlice, response)
	}
	data, err := json.Marshal(returnSlice)
	if err != nil {
		log.Printf("Error marshalling chirps: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (c *APIConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	strID := r.PathValue("chirpID")
	id, err := uuid.Parse(strID)
	if err != nil {
		log.Printf("Error parsing id: %v", err)
	}
	chirp, err := c.database.GetChirp(r.Context(), id)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Chirp not found"))
	} else {
		type successResponse struct {
			Id         uuid.UUID `json:"id"`
			Created_at time.Time `json:"created_at"`
			Updated_at time.Time `json:"updated_at"`
			Body       string    `json:"body"`
			User_id    uuid.UUID `json:"user_id"`
		}
		response := successResponse{
			Id:         chirp.ID,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			Body:       chirp.Body,
			User_id:    chirp.UserID,
		}
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshalling chirp: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)
	}
}

func (c *APIConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type email struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var e email
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&e)
	if err != nil {
		log.Println("Invalid request body")
	}
	user, err := c.database.GetUser(r.Context(), e.Email)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
	} else {
		authToken, err := auth.MakeJWT(user.ID, c.SecretKey, time.Hour)
		ok, err := auth.CheckPasswordHash(e.Password, user.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte("Incorrect email or password"))
		} else if ok {
			type successResponse struct {
				Id            uuid.UUID `json:"id"`
				Created_at    time.Time `json:"created_at"`
				Updated_at    time.Time `json:"updated_at"`
				Email         string    `json:"email"`
				Is_chirpy_red bool      `json:"is_chirpy_red"`
				Token         string    `json:"token"`
				Refresh       string    `json:"refresh_token"`
			}
			response := successResponse{
				Id:            user.ID,
				Created_at:    user.CreatedAt,
				Updated_at:    user.UpdatedAt,
				Email:         user.Email,
				Is_chirpy_red: user.IsChirpyRed,
				Token:         authToken,
				Refresh:       auth.MakeRefreshToken(),
			}
			_, err := c.database.CreateRefresh(r.Context(), database.CreateRefreshParams{
				Token:     response.Refresh,
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
			})
			if err != nil {
				log.Printf("Error creating refresh token: %v", err)
			} else {
				data, err := json.Marshal(response)
				if err != nil {
					log.Printf("Error marshalling JSON: %s", err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				w.Write(data)
			}
		} else {
			w.WriteHeader(401)
			w.Write([]byte("Incorrect email or password"))
		}
	}

}

func (c *APIConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	authHeader, err := auth.GetBearerToken(r.Header)
	if (authHeader == "") || (err != nil) {
		w.WriteHeader(401)
		w.Write([]byte("Missing or invalid Authorization header"))
		return
	}
	refreshToken, err := c.database.GetRefresh(r.Context(), authHeader)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Invalid refresh token"))
		return
	}
	if refreshToken.RevokedAt.Valid {
		w.WriteHeader(401)
		w.Write([]byte("Refresh token has been revoked"))
		return
	}
	if refreshToken.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(401)
		w.Write([]byte("Refresh token has expired"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	newToken, err := auth.MakeJWT(refreshToken.UserID, c.SecretKey, time.Hour)
	if err != nil {
		log.Printf("Error creating new JWT: %v", err)
	}
	type successResponse struct {
		Token string `json:"token"`
	}
	response := successResponse{
		Token: newToken,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
	}
	w.Write(data)
}

func (c *APIConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	authHeader, err := auth.GetBearerToken(r.Header)
	if (authHeader == "") || (err != nil) {
		w.WriteHeader(401)
		w.Write([]byte("Missing or invalid Authorization header"))
		return
	}
	err = c.database.Revoke(r.Context(), authHeader)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Invalid refresh token"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)
}

func (c *APIConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	authHeader, err := auth.GetBearerToken(r.Header)
	if (authHeader == "") || (err != nil) {
		w.WriteHeader(401)
		w.Write([]byte("Missing or invalid Authorization header"))
		return
	}
	token, err := auth.ValidateJWT(authHeader, c.SecretKey)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Invalid token"))
		return
	}
	type email struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var e email
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&e)
	if err != nil {
		log.Println("Invalid request body")
	}
	pass, err := auth.HashPassword(e.Password)
	if err != nil {
		log.Println("Failed hashing password")
	}
	user, err := c.database.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          e.Email,
		HashedPassword: pass,
		ID:             token,
	})
	if err != nil {
		log.Println("Failed updating user")
	}
	type successResponse struct {
		Id            uuid.UUID `json:"id"`
		Created_at    time.Time `json:"created_at"`
		Updated_at    time.Time `json:"updated_at"`
		Email         string    `json:"email"`
		Is_chirpy_red bool      `json:"is_chirpy_red"`
	}
	response := successResponse{
		Id: user.ID, Created_at: user.CreatedAt, Updated_at: user.UpdatedAt, Email: user.Email, Is_chirpy_red: user.IsChirpyRed,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (c *APIConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	authHeader, err := auth.GetBearerToken(r.Header)
	if (authHeader == "") || (err != nil) {
		w.WriteHeader(401)
		w.Write([]byte("Missing Authorization header"))
		return
	}
	id, err := auth.ValidateJWT(authHeader, c.SecretKey)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Invalid token"))
		return
	}
	chirpid, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Error parsing chirpID: %v", err)
		w.WriteHeader(400)
		w.Write([]byte("Invalid chirpID"))
		return
	}
	chirp, err := c.database.GetChirp(r.Context(), chirpid)
	if err != nil {
		log.Printf("Error getting chirp: %v", err)
		w.WriteHeader(404)
		w.Write([]byte("Chirp not found"))
		return
	}
	if chirp.UserID == id {
		err = c.database.DeleteChirp(r.Context(), chirp.ID)
		if err != nil {
			log.Printf("Error deleting chirp: %v", err)
			w.WriteHeader(404)
			w.Write([]byte("Chirp not found"))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(204)
		}
	} else {
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
	}
}

func (c *APIConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {
	authHeader, err := auth.GetAPIKey(r.Header)
	if authHeader != c.APIkey || err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Missing or invalid Authorization header"))
		return
	}

	type data struct {
		Userid string `json:"user_id"`
	}

	type webhook struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	var wbhook webhook
	err = decoder.Decode(&wbhook)
	if err != nil {
		log.Println("Invalid request body")
	}
	if wbhook.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	id, err := uuid.Parse(wbhook.Data.Userid)
	if err != nil {
		log.Printf("Error parsing user_id: %v", err)
		w.WriteHeader(400)
		w.Write([]byte("Invalid user_id"))
		return
	}
	err = c.database.UpgradeUser(r.Context(), id)
	if err != nil {
		log.Printf("Error upgrading user: %v", err)
		w.WriteHeader(404)
		w.Write([]byte("User not found"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	response := errorResponse{Error: message}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, body string, user_id uuid.UUID, config *APIConfig, r *http.Request) {

	chirp, err := config.database.CreateChrip(r.Context(), database.CreateChripParams{
		Body:   body,
		UserID: user_id,
	})
	if err != nil {
		log.Println("Failed creating chirp")
		log.Printf("%v", err)
	}

	type successResponse struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Body       string    `json:"body"`
		User_id    uuid.UUID `json:"user_id"`
	}
	response := successResponse{
		Id:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_id:    chirp.UserID,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func cleanString(s string) string {
	holder := make([]string, len(s))
	capitals := strings.Split(s, " ")
	lower := strings.ToLower(s)
	cleaned := strings.TrimSpace(lower)
	words := strings.Split(cleaned, " ")
	for i, val := range words {
		if val == "kerfuffle" || val == "sharbert" || val == "fornax" {
			holder[i] = "****"
		} else {
			holder[i] = capitals[i]
		}
	}
	penultimate := strings.Join(holder, " ")
	return strings.TrimSpace(penultimate)
}

func Setchirps(c *APIConfig, r *http.Request) []database.Chirp {
	s := r.URL.Query().Get("author_id")
	if s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			log.Printf("Error parsing author_id: %v", err)
		}
		chirps, err := c.database.GetUserChirps(r.Context(), id)
		if err != nil {
			log.Printf("Error getting chirps: %v", err)
		}
		return chirps
	} else {
		chirps, err := c.database.GetChirps(r.Context())
		if err != nil {
			log.Printf("Error getting chirps: %v", err)
		}
		return chirps
	}
}
