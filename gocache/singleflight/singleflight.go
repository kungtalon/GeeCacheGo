package singleflight

import "sync"

// call represents all requests in progress or finished
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group manages all calls with different key
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do makes sure that fn will be called only once for each key at a time
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		// the call is in progress
		g.mu.Unlock()
		// wait for it to finish
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	// when all the other blocked goroutines get the result
	// delete the result of the call
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
