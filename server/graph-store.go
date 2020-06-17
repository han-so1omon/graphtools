package server

import (
	"github.com/han-so1omon/graphtools/structures"
)

type GraphManagerStore interface {
	Insert(structures.GraphDisplayManager)
	GetGraphManager(int) *structures.GraphDisplayManager
}
