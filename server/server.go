package server

import (
	"context"
	//"encoding/json"
	//"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/han-so1omon/graphtools/structures"
)

// Instruction is the format for a request from client to server
type Instruction struct {
	Structure string                 `json:"structure"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params"`
}

const (
	// ServerErrorType denotes a server error on a websocket message
	ServerErrorType = "server-error"
)

type internalError struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

func (e internalError) Error() string {
	return e.Msg
}

func sendInternalError(ctx context.Context, ws *websocket.Conn, err error) {
	log.Println(err)
	wsjson.Write(ctx, ws, err)
}

func sendGraph(ctx context.Context, ws *websocket.Conn, g *structures.GraphDisplayManager) {
	(*g).Lock()
	defer (*g).Unlock()
	log.Println("Sending graph")
	err := wsjson.Write(ctx, ws, *g)
	if err != nil {
		log.Println("sendgraph:", err)
	}
}

func handleInstruction(
	ctx context.Context,
	cancel context.CancelFunc,
	ws *websocket.Conn,
	instruction Instruction,
	g *structures.GraphDisplayManager) {
	var err error

	log.Println(instruction)
	// Execute functions based off of instruction
	// For functions that require locking the graph display manager, unlock immeadiately
	// so that other routines can deal with tree once OnUpdate is called
	if instruction.Structure == structures.RBTreeType {
		switch instruction.Action {
		case "New":
			//TODO maybe handle this with a call to the memory store
			// Mark the existing graph display manager as done so that other
			// operations waiting on this graph display manager can proceed
			if g != nil && *g != nil {
				(*g).Done()
			}

			// Instantiate new structure
			*g = structures.NewRBTree(ctx, cancel)
		case "Insert":
			t := (*g).(*structures.RBTree)
			n, err := t.NewNode(structures.DataNodeTag)
			if err != nil {
				log.Println("Error getting new node for insertion into tree: ")
				return
			}
			err = t.Insert(t.Root, n)
			if err != nil {
				log.Println("Error inserting into tree: ", err)
				return
			}
		case "Delete":
			t := (*g).(*structures.RBTree)
			//log.Println(t.Graph)
			err = t.Delete(t.Root)
			if err != nil {
				log.Println("Error deleting from tree: ", err)
				return
			}
			//log.Println(t.Graph)
		}
	} else if instruction.Structure == structures.GenericGraphManagerType {
	}
	if g != nil {
		(*g).OnUpdate()
	} else {
		sendInternalError(
			ctx,
			ws,
			internalError{ServerErrorType, "Attempting to call update on nil graph"},
		)
	}
}

func receiveInstructions(
	ctx context.Context,
	cancel context.CancelFunc,
	ws *websocket.Conn,
	g *structures.GraphDisplayManager,
) {
	log.Println("Ready to receive instructions")
	done := false
	newInstruction := false
	for !done {
		var instruction Instruction
		if err := wsjson.Read(ctx, ws, &instruction); err != nil {
			if websocket.CloseStatus(err) != -1 {
				log.Println("Websocket closed. Exiting")
				done = true
				newInstruction = false
				cancel()
			} else {
				log.Println("Incorrect instruction format")
				log.Println(err)
				instruction = Instruction{} // Clear instruction
				newInstruction = false
				sendInternalError(
					ctx,
					ws,
					internalError{ServerErrorType, "Incorrect instruction format"},
				)
			}
		} else {
			newInstruction = true
		}
		if newInstruction {
			go handleInstruction(ctx, cancel, ws, instruction, g)
		}
		select {
		case <-ctx.Done():
			done = true
		default:
			continue
		}
	}
	log.Println("Done receiving instructions")
}

// GraphConnect establishes a websocket connection between the graph server and
// the graph display client
func GraphConnect(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Establishing websocket connection")
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:3000"},
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close(websocket.StatusInternalError, "closing because something happened!")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Get pointer to graph manager
	gid := 0
	g := graphManagerStore.GetGraphManager(gid)
	go receiveInstructions(ctx, cancel, ws, g)

	//g := structures.RandomUnidirectionalGraph(100, 100, 100, 200, 125)
	// Load graph

	// Send graph updates until done
	done := false
	for !done {
		if (*g) != nil {
			select {
			case <-ctx.Done():
				done = true
			case <-(*g).Updated():
				go sendGraph(ctx, ws, g)
			}
		} else {
			select {
			case <-ctx.Done():
				done = true
			default:
				continue
			}
		}
	}
	log.Println("Done with connection")
}

var graphManagerStore GraphManagerStore

func NewRouter(store GraphManagerStore) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", GraphConnect)

	graphManagerStore = store

	return router
}
