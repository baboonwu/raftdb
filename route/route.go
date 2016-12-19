package route

import (
	"fmt"
	"log"
	"net/http"
	"raftdb/raftnetwork"

	"github.com/ant0ine/go-json-rest/rest"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

// StartHttpServer 提供 Http 访问接口
func StartHttpServer(port int) {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		rest.Get("/data/:key", GetData),
		rest.Put("/data", PutData),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	strPort := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(strPort, api.MakeHandler()))
}

// GetData
func GetData(w rest.ResponseWriter, r *rest.Request) {
	key := r.PathParam("key")
	log.Printf("key=%v \n", key)

	log.Printf("lead status=%v \n", raftnetwork.RaftNode.Status().SoftState)

	w.WriteJson(key)
}

type KVPair struct {
	Key string
	Val string
}

// PutData
func PutData(w rest.ResponseWriter, r *rest.Request) {

	// parse parameters
	pair := KVPair{}
	err := r.DecodeJsonPayload(&pair)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("key=%v, val=%v \n", pair.Key, pair.Val)

	w.WriteJson(pair)
}
