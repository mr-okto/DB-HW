package writers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Message struct {
	Message string `json:"message"`
}

func WriteError(res http.ResponseWriter, errCode int, errMsg string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(errCode)
	marshalBody, err := json.Marshal(Message{Message: errMsg})
	if err != nil {
		log.Print(err)
		return
	}
	_, _ = res.Write(marshalBody)
}

func WriteEasyJson(res http.ResponseWriter, code int,
	body interface{ MarshalJSON() ([]byte, error) }) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)
	blob, err := body.MarshalJSON()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, _ = res.Write(blob)
}
