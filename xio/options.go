package xio

type copyoptions struct {
	WaitForLastOp bool
	bufferSize    int
	buffer        []byte
}

type CopyOption func(*copyoptions)

func WaitForLastOp(value bool) CopyOption {
	return func(c *copyoptions) {
		c.WaitForLastOp = value
	}
}

func BufferSize(value int) CopyOption {
	return func(c *copyoptions) {
		c.bufferSize = value
	}
}

func Buffer(b []byte) CopyOption {
	return func(c *copyoptions) {
		c.buffer = b
	}
}
