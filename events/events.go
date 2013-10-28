package events

import "strconv"

type Event interface {
	EventName() string
}

type EventMessage struct {
	Content Event
	Source  EventSource
}

func (msg EventMessage) EventName() string {
	return "EventMessage<" + msg.Content.EventName() + ">"
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
	EventType  string
	Source     EventSource
}

func (s Subscription) EventName() string {
	return "Subscription<" + s.EventType + ">"
}

type Cancellation struct {
	Subscription
}

func (s Cancellation) EventName() string {
	return "Cancellation<" + s.EventType + ">"
}

type EventQueue struct {
	Relay       EventChannel
	Subscribers map[Subscription]EventSource
}

func NewEventQueue() *EventQueue {
	q := new(EventQueue)
	q.Relay = make(EventChannel)
	q.Subscribers = make(map[Subscription]EventSource, 1)
	go q.relayEvents()
	return q
}

func (queue *EventQueue) relayEvents() {
	for event := range queue.Relay {
		// relay events
		switch t := event.(type) {
		case Subscription:
			// adds a new subscriber to this queue
			queue.Subscribers[t] = t.Subscriber
		case Cancellation:
			// removes a subscriber from this queue
			delete(queue.Subscribers, t.Subscription)
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
func QueueName(event string, src EventSource) string {
	if src != nil {
		return event + "@" + strconv.FormatUint(src.EventSourceID(), 32)
	} else {
		return event
	}
}

//-----------------------------------------------------------------------------
// EVENT BUS

type EventTransmitter interface {
	// Subscribes to events of given type and optionally (source > 0) to events
	// triggered by a specific other entity only
	Subscribe(subscriber EventSource, event string, source EventSource)
	Unsubscribe(subscriber EventSource, event string, source EventSource)

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
	bus.subscribeRequests = make(chan Subscription)
	bus.unsubscribeRequests = make(chan Cancellation)
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

func (bus *EventBus) Subscribe(subscriber EventSource, event string, source EventSource) {
	bus.subscribeRequests <- Subscription{subscriber, event, source}
}

func (bus *EventBus) Unsubscribe(subscriber EventSource, event string, source EventSource) {
	bus.unsubscribeRequests <- Cancellation{Subscription{subscriber, event, source}}
}

func (bus *EventBus) Trigger(event Event, source EventSource) {
	bus.eventRequests <- EventMessage{event, source}
}

func (bus *EventBus) Run() {
	for bus.queues != nil {
		select {
		case subscription, ok := <-bus.subscribeRequests:
			if ok {
				queue := getQueueOrCreate(bus, subscription.EventType, subscription.Source)
				queue.Relay <- subscription
			}
		case cancellation, ok := <-bus.unsubscribeRequests:
			if ok {
				queue := getQueueOrCreate(bus, cancellation.EventType, cancellation.Source)
				queue.Relay <- cancellation
			}
		case event, ok := <-bus.eventRequests:
			if ok {
				if generalQueue, ok := getQueue(bus, event.EventName(), nil); ok {
					generalQueue.Relay <- event
				}

				if specificQueue, ok := getQueue(bus, event.EventName(), event.Source); ok {
					specificQueue.Relay <- event
				}
			}
		}
	}
}

func getQueue(bus *EventBus, event string, source EventSource) (*EventQueue, bool) {
	queueName := QueueName(event, source)
	queue, ok := bus.queues[queueName]
	if !ok {
		return nil, false
	}

	return queue, true
}

func getQueueOrCreate(bus *EventBus, event string, source EventSource) *EventQueue {
	queueName := QueueName(event, source)
	queue, ok := bus.queues[queueName]
	if !ok {
		queue = NewEventQueue()
		bus.queues[queueName] = queue
	}

	return queue
}
