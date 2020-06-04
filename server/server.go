package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"

	"github.com/han-so1omon/graphtools/structures"
)

var upgrader = websocket.Upgrader{}

func internalError(ws *websocket.Conn, msg string, err error) {
	log.Println(msg, err)
	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
}

func sendGraph(ws *websocket.Conn, g *structures.Graph) {
	g.Lock.Lock()
	gJSON, err := json.Marshal(g)
	if err != nil {
		log.Println("sendgraph:", err)
	}
	ws.WriteMessage(websocket.TextMessage, gJSON)
	g.Lock.Unlock()
}

func GraphConnect(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Setup websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	// Load graph
	g := structures.RandomBidirectionalGraph(100, 100, 100, 200, 125)

	// Send initial graph
	go sendGraph(ws, g)

	defer ws.Close()

	// Send graph updates until done
	done := false
	for {
		select {
		case <-g.Done:
			done = true
		case <-g.Updated:
			go sendGraph(ws, g)
		}
		if done {
			break
		}
	}
}

func NewRouter() *httprouter.Router {
	router := httprouter.New()
	router.GET("/ws", GraphConnect)

	return router
}
