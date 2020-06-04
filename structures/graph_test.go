package structures

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestRandomGraph(t *testing.T) {
	g := RandomBidirectionalGraph(10, 10, 100, 20, 20)
	fmt.Println(g)
}
