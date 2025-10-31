//go:build !dev

package dummy

type counts struct{}

func (c *counts) acceptCountAdd()  {}
func (c *counts) workerCountAdd()  {}
func (c *counts) processCountAdd() {}
func (c *counts) requestCountAdd() {}
