package producer

import (
	"encoding/json"
	webhook "github.com/xmidt-org/wrp-listener"
	"github.com/xmidt-org/wrp-listener/producer/forwarder"
	"io/ioutil"
	"net/http"
)

type WebhookServer struct {
	Forwader *forwarder.Forwader
}

func (app *WebhookServer) RegisterListener(writer http.ResponseWriter, req *http.Request) {
	payload, err := ioutil.ReadAll(req.Body)
	req.Body.Close()

	w, err := webhook.NewW(payload, req.RemoteAddr)
	if err != nil {
		panic(err)
		writer.Header().Add("X-Xmidt-Error", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = app.Forwader.Update(*w)
	if err != nil {
		panic(err)

		writer.Header().Add("X-Xmidt-Error", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (app *WebhookServer) GetSanatizedWebhooks(writer http.ResponseWriter, req *http.Request) {
	hooks := app.Forwader.GetHooks()
	data, err := json.Marshal(&hooks)
	if err != nil {
		panic(err)

		writer.Header().Add("X-Xmidt-Error", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: sanatize input
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(data)
	writer.WriteHeader(http.StatusOK)
}
