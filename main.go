package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"google.golang.org/api/chat/v1"

	"golang.org/x/oauth2/google"
	"google.golang.org/appengine" // Required external App Engine library
	"google.golang.org/appengine/log"
)

type CircleCIPayload struct {
	Payload CircleCI `json:"payload"`
}

type CircleCI struct {
	VCSURL          string `json:"vcs_url"`
	BuildURL        string `json:"build_url"`
	BuildNumber     int    `json:"build_num"`
	Branch          string `json:"branch"`
	VCSRevision     string `json:"`
	CommitterName   string `json:"committer_name"`
	CommiterEmail   string `json:"committer_email"`
	Subject         string `json:"subject"`
	Body            string `json:"body"`
	Why             string `json:"why"`
	BuildTimeMillis int    `json:"build_time_millis"`
	DontBuild       string `json:"dont_build"`
	QueuedAt        string `json:"queue_at"`
	StartTime       string `json:"start_time"`
	StopTime        string `json:"stop_time"`
	UserName        string `json:"username"`
	RepoName        string `json:"reponame"`
	Lifecycle       string `json:"lifecycle"`
	Outcome         string `json:"outcome"`
	Status          string `json:"status"`
}

var err error

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Set Headers
	ctx := appengine.NewContext(r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	token, ok := r.URL.Query()["token"]

	if !ok || token[0] != os.Getenv("TOKEN") {
		log.Infof(ctx, "No Param Received")
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Infof(ctx, "Error Reading Body", err)
		// json.NewEncoder(w).Encode(e)
		return
	}
	log.Infof(ctx, "Body %v", string(b))
	var msg CircleCIPayload
	err = json.Unmarshal(b, &msg)
	if err != nil {
		log.Infof(ctx, "Error :%v", err)
	}
	text := "Build number " + strconv.Itoa(msg.Payload.BuildNumber) + " on Repository: " + msg.Payload.Branch + "/" + msg.Payload.RepoName + ", has completed with the status of: " + msg.Payload.Status + ". The build was kicked off by " + msg.Payload.CommitterName + ". The build took " + strconv.Itoa(msg.Payload.BuildTimeMillis/1000) + " seconds. More information is available here: " + msg.Payload.BuildURL
	client, err := google.DefaultClient(ctx,
		"https://www.googleapis.com/auth/chat.bot")
	if err != nil {
		log.Infof(ctx, "Error: %v", err)
	}
	response := chat.Message{Text: text, Sender: &chat.User{DisplayName: "Hubot"}}
	postBody, err := json.Marshal(response)
	if err != nil {
		log.Infof(ctx, "Error Converting Response to Reader")
		return
	}
	client.Post("https://chat.googleapis.com/v1/spaces/AAAAV2Ons90/messages?threadKey=build", "application/json", bytes.NewBuffer(postBody))
	json.NewEncoder(w).Encode("Success")
	return
}

func main() {
	http.HandleFunc("/", indexHandler)
	appengine.Main() // Starts the server to receive requests
}
