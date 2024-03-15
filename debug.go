package remotedialer

import (
	"encoding/json"
	"net/http"
	"sync"
)

type smInfo struct {
	sms []*sessionManager
	*sync.Mutex
}

var info = &smInfo{
	Mutex: &sync.Mutex{},
}

func init() {
	http.HandleFunc("/debug/session-manager", func(w http.ResponseWriter, r *http.Request) {
		type sessionManagerData struct {
			Clients []string            `json:"clients"`
			Peers   map[string][]string `json:"peers"`
		}

		info.Lock()
		sms := info.sms
		info.Unlock()

		sessionManagers := make([]sessionManagerData, len(sms))

		for i, sm := range sms {
			data := &sessionManagers[i]
			sm.Lock()
			data.Clients = make([]string, 0, len(sm.clients))
			for client := range sm.clients {
				data.Clients = append(data.Clients, client)
			}
			data.Peers = map[string][]string{}
			for name, sessions := range sm.peers {
				for _, session := range sessions {
					session.Lock()
					for clientKey := range session.remoteClientKeys {
						data.Peers[name] = append(data.Peers[name], clientKey)
					}
					session.Unlock()
				}
			}
			sm.Unlock()
		}

		out, err := json.Marshal(sessionManagers)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Write(out)
	})
}
