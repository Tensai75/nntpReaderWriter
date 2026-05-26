package nntpReaderWriter

import (
	"sync"
	"sync/atomic"
)

type sequencer struct {
	id   atomic.Uint64
	mu   sync.Mutex
	seq  uint
	wait map[uint]chan struct{}
}

// Next returns the next sequencer id.
func (p *sequencer) Next() uint {
	return uint(p.id.Add(1) - 1)
}

// Start blocks until the sequencer is ready for the given id.
func (p *sequencer) Start(id uint) {
	p.mu.Lock()
	if p.seq == id {
		p.mu.Unlock()
		return
	}
	c := make(chan struct{})
	if p.wait == nil {
		p.wait = make(map[uint]chan struct{})
	}
	p.wait[id] = c
	p.mu.Unlock()
	<-c
}

// End signals that the sequencer for the given id is complete and unblocks the next in sequence.
func (p *sequencer) End(id uint) {
	p.mu.Lock()
	if p.seq != id {
		p.mu.Unlock()
		// This should never happen if the sequencer is used correctly
		// but we can panic here to catch any bugs in our usage of the sequencer.
		panic("sequencer out of sync")
	}
	id++
	p.seq = id
	if p.wait == nil {
		p.wait = make(map[uint]chan struct{})
	}
	c, ok := p.wait[id]
	if ok {
		delete(p.wait, id)
	}
	p.mu.Unlock()
	if ok {
		close(c)
	}
}
