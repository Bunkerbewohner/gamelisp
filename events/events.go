package events

import "reflect"
import "strconv"

type Event interface {
}

type EventMessage struct {
	Content Event
	Source  EventSource
}

type EventSource interface {
	// Channel that triggered events are send into
	EventChannel() EventChannel

	// Unique Identifier for this event source
	EventSourceID() uint64
}

type EventChannel chan Event

type Subscription struct {
	Subscriber EventSource
	EventType  reflect.Type
	Source     EventSource
}

type Cancellation struct {
	Subscription
}

type EventQueue struct {
	Relay       EventChannel
	Subscribers map[uint64]EventSource
}

func NewEventQueue() *EventQueue {
	q := new(EventQueue)
	q.Relay = make(EventChannel)
	q.Subscribers = make(map[uint64]EventSource, 1)
	go q.relayEvents()
	return q
}

func (queue *EventQueue) relayEvents() {
	for event := range queue.Relay {
		switch t := event.(type) {
		case Subscription:
			// adds a new subscriber to this queue
			queue.Subscribers[t.Subscriber.EventSourceID()] = t.Subscriber
		case Cancellation:
			// removes a subscriber from this queue
			delete(queue.Subscribers, t.Subscriber.EventSourceID())
		default:
			// relays this event to all subscribers
			for _, subscriber := range queue.Subscribers {
				subscriber.EventChannel() <- event
			}
		}
	}
}

// QueueName returns a canonical name for the channel that should be used
// for events of the given type and for the specified event source.
func QueueName(event reflect.Type, src EventSource) string {
	if src != nil {
		return event.PkgPath() + event.Name() + "@" + strconv.FormatUint(src.EventSourceID(), 32)
	} else {
		return event.PkgPath() + event.Name()
	}
}

//-----------------------------------------------------------------------------
// EVENT BUS

type EventTransmitter interface {
	// Subscribes to events of given type and optionally (source > 0) to events
	// triggered by a specific other entity only
	Subscribe(subscriber EventSource, event reflect.Type, source EventSource)
	Unsubscribe(subscriber EventSource, event reflect.Type, source EventSource)

	// Triggers an event providing the ID of the originating entity
	Trigger(event Event, source EventSource)
}

type EventBus struct {
	subscribeRequests   chan Subscription
	unsubscribeRequests chan Cancellation
	eventRequests       chan EventMessage
	queues              map[string]*EventQueue
}

func (bus *EventBus) Init() {
	bus.subscribeRequests = make(chan Subscription, 10)
	bus.unsubscribeRequests = make(chan Cancellation, 10)
	bus.eventRequests = make(chan EventMessage, 1000)
	bus.queues = make(map[string]*EventQueue, 100)

	go bus.Run()
}

func (bus *EventBus) Shutdown() {
	for _, queue := range bus.queues {
		close(queue.Relay)
	}

	bus.queues = nil
	close(bus.subscribeRequests)
	close(bus.unsubscribeRequests)
	close(bus.eventRequests)
}

func (bus *EventBus) Subscribe(subscriber EventSource, event reflect.Type, source EventSource) {
	bus.subscribeRequests <- Subscription{subscriber, event, source}
}

func (bus *EventBus) Unsubscribe(subscriber EventSource, event reflect.Type, source EventSource) {
	bus.unsubscribeRequests <- Cancellation{Subscription{subscriber, event, source}}
}

func (bus *EventBus) Trigger(event Event, source EventSource) {
	bus.eventRequests <- EventMessage{event, source}
}

func (bus *EventBus) Run() {
	for bus.queues != nil {
		select {
		case subscription := <-bus.subscribeRequests:
			queue := getQueueOrCreate(bus, subscription.EventType, subscription.Source)
			queue.Relay <- subscription
		case cancellation := <-bus.unsubscribeRequests:
			queue := getQueueOrCreate(bus, cancellation.EventType, cancellation.Source)
			queue.Relay <- cancellation
		case event := <-bus.eventRequests:
			if generalQueue, ok := getQueue(bus, reflect.TypeOf(event.Content), nil); ok {
				generalQueue.Relay <- event
			}

			if specificQueue, ok := getQueue(bus, reflect.TypeOf(event.Content), event.Source); ok {
				specificQueue.Relay <- event
			}
		}
	}
}

func getQueue(bus *EventBus, event reflect.Type, source EventSource) (*EventQueue, bool) {
	queueName := QueueName(event, source)
	queue, ok := bus.queues[queueName]
	if !ok {
		return nil, false
	}

	return queue, true
}

func getQueueOrCreate(bus *EventBus, event reflect.Type, source EventSource) *EventQueue {
	queueName := QueueName(event, source)
	queue, ok := bus.queues[queueName]
	if !ok {
		queue = NewEventQueue()
		bus.queues[queueName] = queue
	}

	return queue
}

func transmitEvents(bus *EventBus, channel EventChannel) {

}

//-----------------------------------------------------------------------------
