package server

import (
	"github.com/han-so1omon/graphtools/structures"
)

type InMemoryGraphStore struct {
	GraphManager structures.GraphDisplayManager
}

func (s *InMemoryGraphStore) Insert(mgr structures.GraphDisplayManager) {
	s.GraphManager = mgr
}

func (s *InMemoryGraphStore) GetGraphManager(id int) *structures.GraphDisplayManager {
	return &s.GraphManager
}
