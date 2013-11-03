package main

import gl "github.com/go-gl/gl"
import "sync"

type GraphicsQueue struct {
	calls []func()
}

var graphicsQueue = NewGraphicsQueue()
var lock = new(sync.Mutex)

func NewGraphicsQueue() *GraphicsQueue {
	queue := new(GraphicsQueue)
	queue.calls = make([]func(), 0, 100)

	return queue
}

func (gq *GraphicsQueue) Process() {
	calls := make([]func(), len(gq.calls))
	lock.Lock()
	copy(calls, gq.calls)
	gq.calls = make([]func(), 0, 100)
	lock.Unlock()

	for _, call := range calls {
		call()
	}
}

func (gq *GraphicsQueue) Enqueue(call func()) {
	lock.Lock()
	gq.calls = append(gq.calls, call)
	lock.Unlock()
}

func fill_background(args List, context *Context) Data {
	red := args.First().(Float)
	green := args.Second().(Float)
	blue := args.Third().(Float)
	alpha := args.Get(3).(Float)

	graphicsQueue.Enqueue(func() {
		gl.ClearColor(gl.GLclampf(red.Value), gl.GLclampf(green.Value),
			gl.GLclampf(blue.Value), gl.GLclampf(alpha.Value))
	})

	return Nothing{}
}
