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

func (e *Entity) Set(property string, value Data) {
	if old, ok := e.data[property]; ok {
		e.data[property] = value
		// do sth with old
		fmt.Printf(old.String())
	} else {
		e.data[property] = value
	}
}

var shutdown = false
var newEntities = make(chan *Entity)
var ids = make(chan uint64)

func ECS_init() {
	go generate_entity_ids(newEntities, ids)
}

func ECS_shutdown() {
	shutdown = true
	close(newEntities)
}

func NewEntity() *Entity {
	ent := new(Entity)
	newEntities <- ent
	ent.id = <-ids
	return ent
}

func generate_entity_ids(entities chan *Entity, ids chan uint64) {
	i := uint64(1)
	for e := range entities {
		if e.id == 0 {
			ids <- i
		}
		i++
	}
	close(ids)
}
