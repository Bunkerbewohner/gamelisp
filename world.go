package main

import "container/list"

type World struct {
	// preallocated pool of cubes
	cubePool []Cube

	// index of next allocatable cube in pool
	poolAllocIndex int

	// List of cubes that had been used before and are now free to use again
	freedCubes *list.List
}

func NewWorld() *World {
	world := new(World)
	world.cubePool = make([]Cube, 10000)
	world.freedCubes = list.New()

	return world
}

func (world *World) allocateCube() *Cube {
	if world.freedCubes.Len() > 0 {
		instance := world.freedCubes.Front()
		world.freedCubes.Remove(instance)

		index := instance.Value.(int)
		return &world.cubePool[index]
	}

	if world.poolAllocIndex >= len(world.cubePool) {
		world.cubePool = append(world.cubePool, Cube{})
	}

	world.poolAllocIndex++
	cube := &world.cubePool[world.poolAllocIndex-1]
	cube.used = true
	cube.index = world.poolAllocIndex - 1

	return cube
}

func (world *World) Create(cube Cube) {
	instance := world.allocateCube()
	*instance = cube
	instance.used = true
}

func (world *World) Destroy(cubeIndex int) {
	world.cubePool[cubeIndex].used = false
	world.freedCubes.PushBack(cubeIndex)
}

func (world *World) FindCubeAt(pos Vertex3D) *Cube {
	// TODO: Optimize FindCubeAt
	for _, cube := range world.cubePool {
		if cube.used && cube.Position.Similar(pos) {
			return &world.cubePool[cube.index]
		}
	}

	return nil
}

type Cube struct {
	index int
	used  bool

	Position Vertex3D
	Size     float64
}

func CreateCube(x, y, z float64) Cube {
	return Cube{
		Position: Vertex3D{x, y, z},
	}
}
