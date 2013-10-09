package main

import "fmt"

/* Entity Component System */

type Entity struct {
	id   uint64
	data map[string]Data
}

func (e *Entity) String() string {
	return fmt.Sprintf("Entity<%d>", e.id)
}

func (e *Entity) GetType() DataType {
	return EntityType
}

func (e *Entity) Equals(other Data) bool {
	switch t := other.(type) {
	case *Entity:
		return t.id == e.id
	}
	return false
}

type Trait struct {
}

var shutdown = false
var newEntities = make(chan *Entity)

func ECS_init() {
	go generate_entity_ids(newEntities)
}

func ECS_shutdown() {
	shutdown = true
	close(newEntities)
}

func NewEntity() *Entity {
	ent := new(Entity)
	newEntities <- ent
	return ent
}

func generate_entity_ids(entities chan *Entity) {
	i := uint64(1)
	for e := range entities {
		e.id = i
		i++
	}
}
