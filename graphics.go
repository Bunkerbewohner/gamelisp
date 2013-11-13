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

//--------------------------------
// Cube / World Rendering

type Vertex3D struct {
	X, Y, Z float64
}

func (v Vertex3D) Similar(other Vertex3D) bool {
	dx := other.X - v.X
	dy := other.Y - v.Y
	dz := other.Z - v.Z

	return (dx*dx + dy*dy + dz*dz) < 0.5
}

type Vertex4D struct {
	Vertex3D
	W float64
}

func (cube *Cube) Render() {
	x, y, z := cube.Position.X, cube.Position.Y, cube.Position.Z

	gl.Begin(gl.QUADS)

	gl.Color3d(0.5-(x/50), 0.5-(y/50), 0.5-(z/50))

	// Front Side
	gl.TexCoord2f(0, 0)
	gl.Vertex3d(x-0.5, y-0.5, z+0.5)
	gl.TexCoord2f(1, 0)
	gl.Vertex3d(x+0.5, y-0.5, z+0.5)
	gl.TexCoord2f(1, 1)
	gl.Vertex3d(x+0.5, y+0.5, z+0.5)
	gl.TexCoord2f(0, 1)
	gl.Vertex3d(x-0.5, y+0.5, z+0.5)

	// Left Side
	gl.Color3d(0.5-(x/20), 0.5-(y/20), 0.5-(z/20))
	gl.TexCoord2f(0, 0)
	gl.Vertex3d(x-0.5, y-0.5, z-0.5)
	gl.TexCoord2f(1, 0)
	gl.Vertex3d(x-0.5, y-0.5, z+0.5)
	gl.TexCoord2f(1, 1)
	gl.Vertex3d(x-0.5, y+0.5, z+0.5)
	gl.TexCoord2f(0, 1)
	gl.Vertex3d(x-0.5, y+0.5, z-0.5)

	gl.End()
}

func (world *World) Render() {
	for _, cube := range world.cubePool {
		if !cube.used {
			continue
		}

		cube.Render()
	}
}
