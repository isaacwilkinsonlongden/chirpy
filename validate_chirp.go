package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/isaacwilkinsonlongden/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong", err)
		return
	}

	cleanedBody, ok := validateChirp(params.Body)
	if !ok {
		respondWithError(w, 400, "Chirp is too long", nil)
		return
	}
	params.Body = cleanedBody

	chirpParams := database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, 500, "Something went wrong", err)
		return
	}

	respondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Error fetching chirps", err)
		return
	}

	returnChirps := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		returnChirps[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	respondWithJSON(w, 200, returnChirps)
}

func validateChirp(chirp string) (string, bool) {
	if len(chirp) > 140 {
		return "", false
	}

	return replaceBadWords(chirp), true
}

func replaceBadWords(body string) string {
	wordsSlice := strings.Split(body, " ")
	cleanedWordsSlice := make([]string, len(wordsSlice))

	for i, word := range wordsSlice {
		lowercaseWord := strings.ToLower(word)
		if lowercaseWord == "kerfuffle" || lowercaseWord == "sharbert" || lowercaseWord == "fornax" {
			cleanedWordsSlice[i] = "****"
		} else {
			cleanedWordsSlice[i] = word
		}
	}

	return strings.Join(cleanedWordsSlice, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}
