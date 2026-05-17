package nntpReaderWriter

import (
	"sync"
	"sync/atomic"
)

type pipeline struct {
	id  atomic.Uint64
	cmd sequencer
}

func (p *pipeline) Next() uint {
	return uint(p.id.Add(1) - 1)
}

func (p *pipeline) Start(id uint) {
	p.cmd.Start(id)
}

func (p *pipeline) End(id uint) {
	p.cmd.End(id)
}

type sequencer struct {
	mu   sync.Mutex
	id   uint
	wait map[uint]chan struct{}
}

func (s *sequencer) Start(id uint) {
	s.mu.Lock()
	if s.id == id {
		s.mu.Unlock()
		return
	}
	c := make(chan struct{})
	if s.wait == nil {
		s.wait = make(map[uint]chan struct{})
	}
	s.wait[id] = c
	s.mu.Unlock()
	<-c
}

func (s *sequencer) End(id uint) {
	s.mu.Lock()
	if s.id != id {
		s.mu.Unlock()
		// This should never happen if the pipeline is used correctly
		// but we can panic here to catch any bugs in our usage of the pipeline.
		panic("sequencer out of sync")
	}
	id++
	s.id = id
	if s.wait == nil {
		s.wait = make(map[uint]chan struct{})
	}
	c, ok := s.wait[id]
	if ok {
		delete(s.wait, id)
	}
	s.mu.Unlock()
	if ok {
		close(c)
	}
}
