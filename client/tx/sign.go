package tx

import (
	"encoding/json"
	"net/http"

	keybase "github.com/tepleton/tepleton-sdk/client/keys"
	keys "github.com/tepleton/go-crypto/keys"
)

type SignTxBody struct {
	Name     string `json="name"`
	Password string `json="password"`
	TxBytes  string `json="tx"`
}

func SignTxRequstHandler(w http.ResponseWriter, r *http.Request) {
	var kb keys.Keybase
	var m SignTxBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	kb, err = keybase.GetKeyBase()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	//TODO check if account exists
	sig, _, err := kb.Sign(m.Name, m.Password, []byte(m.TxBytes))
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(sig.Bytes())
}