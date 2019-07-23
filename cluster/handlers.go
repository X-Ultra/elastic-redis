package cluster

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type joinHandler struct {}
func (h *joinHandler)handleJoin(s *Server){

}

type infoHandler struct {}
func (h *infoHandler) handleInfo(s *Server){

}

type RootHandler struct {
	S *Server
	joinHandler
	infoHandler
}

func (h *RootHandler)ServeHTTP(w http.ResponseWriter, r *http.Request){
	head, path := shiftPath(r.URL.Path)
	fmt.Println("head: ", head, ", path: ", path)

	switch head {
	case "join":
		h.handleJoin(h.S)
	case "info":
		h.handleInfo(h.S)
	default:
		w.WriteHeader(http.StatusNotFound)
	}

}

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}