package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"
)

type AlbumsData struct {
	Name        string
	Image       string
	ReleaseDate string
	Tracks      int
}

type TrackData struct {
	Title       string
	AlbumCover  string
	Album       string
	Artist      string
	ReleaseDate string
	SpotifyLink string
}

func main() {

	temp, err := template.ParseGlob("SpotifyAPI/site_web/web/templates/*")
	if err != nil {
		fmt.Println("Erreur > ", err)
		return
	}

	julurl := "https://api.spotify.com/v1/artists/3IW7ScrzXmPvZhB27hmfgy/albums"
	sdmurl := "https://api.spotify.com/v1/tracks/0EzNyXyU7gHzj2TN8qYThj?market=FR"

	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		temp.ExecuteTemplate(w, "index", nil)
	})

	http.HandleFunc("/album/jul", func(w http.ResponseWriter, r *http.Request) {
		clientID := "9bd2f253c4c04f6ca2eeb009d9127326"
		clientSecret := "91358e737e5245ce8f28b44cc8f9ffc1"

		authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

		token, err := getAccessToken(authHeader)
		if err != nil {
			log.Fatalf("Impossible d'obtenir un token: %v", err)
		}

		req, errReq := http.NewRequest(http.MethodGet, julurl, nil)
		if errReq != nil {
			fmt.Println("Un problème est survenu : ", errReq.Error())
		}

		req.Header.Add("Authorization", "Bearer "+token)

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Un problème est survenu : ", errRes.Error())
		}

		var Albums map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&Albums); err != nil {
			fmt.Println("Un problème est survenu : ", err.Error())
		}

		albumsData, bool := Albums["items"].([]interface{})
		if !bool {
			fmt.Println("Un problème est survenu : aucun album trouvé")
		}

		var decodeData []AlbumsData

		for _, album := range albumsData {
			Data := album.(map[string]interface{})
			var New AlbumsData

			if len(Data["images"].([]interface{})) > 0 {
				New.Image = Data["images"].([]interface{})[0].(map[string]interface{})["url"].(string)
			}
			New.Name = Data["name"].(string)
			New.ReleaseDate = Data["release_date"].(string)
			New.Tracks = int(Data["total_tracks"].(float64))
			decodeData = append(decodeData, New)

		}

		temp.ExecuteTemplate(w, "jul", decodeData)
	})

	http.HandleFunc("/track/sdm", func(w http.ResponseWriter, r *http.Request) {
		clientID := "9bd2f253c4c04f6ca2eeb009d9127326"
		clientSecret := "91358e737e5245ce8f28b44cc8f9ffc1"

		authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

		token, err := getAccessToken(authHeader)
		if err != nil {
			log.Fatalf("Impossible d'obtenir un token: %v", err)
		}

		req, errReq := http.NewRequest(http.MethodGet, sdmurl, nil)
		if errReq != nil {
			fmt.Println("Un problème est survenu : ", errReq.Error())
		}

		req.Header.Add("Authorization", "Bearer "+token)

		res, errRes := httpClient.Do(req)
		if res.Body != nil {
			defer res.Body.Close()
		} else {
			fmt.Println("Un problème est survenu : ", errRes.Error())
		}

		var Track map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&Track); err != nil {
			fmt.Println("Un problème est survenu : ", err.Error())
		}

		Info := TrackData{}
		Info.Title = Track["name"].(string)
		Info.AlbumCover = Track["album"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["url"].(string)
		Info.Album = Track["album"].(map[string]interface{})["name"].(string)
		artists := Track["artists"].([]interface{})
		if len(artists) > 0 {
			Info.Artist = artists[0].(map[string]interface{})["name"].(string)
		}
		Info.ReleaseDate = Track["album"].(map[string]interface{})["release_date"].(string)
		Info.SpotifyLink = Track["external_urls"].(map[string]interface{})["spotify"].(string)

		temp.ExecuteTemplate(w, "sdm", Info)
	})

	http.ListenAndServe("localhost:8080", nil)
}

func getAccessToken(authHeader string) (string, error) {
	tokenURL := "https://accounts.spotify.com/api/token"
	payload := "?grant_type=client_credentials"
	body := tokenURL + payload
	req, err := http.NewRequest("POST", body, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Basic "+authHeader)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("token d'accès introuvable")
	}

	return accessToken, nil
}
