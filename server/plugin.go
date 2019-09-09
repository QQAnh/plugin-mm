package main

import (
	"bytes"
	//"bytes"
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"io/ioutil"

	"github.com/mattermost/mattermost-server/plugin"
	//"fmt"
	//"github.com/mattermost/mattermost-server/model"
	//"io/ioutil"
	"net/http"
	"sync"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/status":
		p.handleStatus(w, r)
	case "/hello":
		p.handleHello(c,w, r)
	default:
		http.NotFound(w, r)
	}
}
func (p *Plugin) handleStatus(w http.ResponseWriter, r *http.Request) {
	//configuration := p.getConfiguration()

	var response = struct {
		Enabled bool `json:"enabled"`
	}{
		//Enabled: !configuration.disabled,
	}

	responseJSON, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(responseJSON); err != nil {
		p.API.LogError("failed to write status", "err", err.Error())
	}
}

func (p *Plugin) handleHello(c *plugin.Context,w http.ResponseWriter, r *http.Request ) {

	props := model.MapFromJson(r.Body)
	email := props["email"]
	username := props["username"]
	password := props["password"]
	//user := model.User{Email:email,Password:password,Username:username}
	if _, appErr := p.API.CreateUser(&model.User{
		Email:    email,
		Username: username,
		Password: password,
	}); appErr == nil {
		//p.API.LogInfo("create new user", "body", a)
		//w.Header().Set("Content-Type", "application/json")
		//w.WriteHeader(http.StatusOK)
		//w.Write([]byte(a.ToJson()))
	} else {
		p.API.LogError("failed to create new user", "err", appErr.Error())
		//w.Header().Set("Content-Type", "application/json")
		//w.WriteHeader(http.StatusOK)
		//w.Write([]byte(appErr.Error()))
		//http.Error(w, appErr.Error(), http.StatusTeapot)
	}

	url := "http://127.0.0.1:8065/api/v4/users/login"
	fmt.Println("URL:>", url)

	var jsonStr = []byte(`{"login_id":"`+ email + `","password":"`  + password +`"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if resp.Status == "200 OK" {
		byte, err := json.Marshal(resp.Header.Get("Set-Cookie"))

		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(byte))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(string(body)))
	return
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
