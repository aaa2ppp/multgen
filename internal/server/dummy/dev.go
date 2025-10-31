//go:build dev

package dummy

import "sync/atomic"

type counts struct {
	acceptCount  atomic.Int64
	workerCount  atomic.Int64
	processCount atomic.Int64
	requestCount atomic.Int64
}

func (c *counts) acceptCountAdd()  { c.acceptCount.Add(1) }
func (c *counts) workerCountAdd()  { c.workerCount.Add(1) }
func (c *counts) processCountAdd() { c.processCount.Add(1) }
func (c *counts) requestCountAdd() { c.requestCount.Add(1) }

func (c *counts) AcceptCount() int64  { return c.acceptCount.Load() }
func (c *counts) WorkerCount() int64  { return c.workerCount.Load() }
func (c *counts) ProcessCount() int64 { return c.processCount.Load() }
func (c *counts) RequestCount() int64 { return c.requestCount.Load() }
