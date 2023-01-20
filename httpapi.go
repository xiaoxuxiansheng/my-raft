package main

import (
	"io"
	"net/http"
	"strconv"
	"xiaoxuxiansheng/my-raft/raft"
)

type service struct {
	proposeC    chan<- string
	confChangeC chan<- raft.ConfChange
	commitC     <-chan *string
}

func newService(proposeC chan<- string, confChangeC chan<- raft.ConfChange, commitC <-chan *string) *service {
	return &service{
		proposeC:    proposeC,
		confChangeC: confChangeC,
		commitC:     commitC,
	}

}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.RequestURI

	switch {
	case r.Method == http.MethodPut:
		v, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		s.proposeC <- string(v)

	case r.Method == http.MethodPost:
		v, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		nodeID, err := strconv.ParseUint(url[1:], 0, 64)
		if err != nil {
			panic(err)
		}
		s.confChangeC <- raft.ConfChange{
			NodeID:  nodeID,
			Type:    raft.ConfChangeAddNode,
			Context: v,
		}

	}

}

func serveHttoAPI(port int, s *service) {
	srv := http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: s,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
