package target

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Topic is a subscription topic.
type Topic int

const (
	// TopicProcessCreated is the subscription topic that all process creation
	// events are sent on.
	TopicProcessCreated Topic = 1
	// TopicProcessRemoved is the subscription topic that all process
	// completions events are sent on.
	TopicProcessRemoved Topic = 2
)

// Broker brokers the distribution of Process state change discovered by a
// source to subscriptions.
type Broker struct {
	logger *slog.Logger

	subs     map[Topic]map[string]*Subscription
	subsLock sync.Mutex
}

// NewBroker creates a new broker instance.
func NewBroker(logger *slog.Logger) *Broker {
	if logger == nil {
		logger = slog.Default()
	}

	return &Broker{
		logger: logger,
		subs: map[Topic]map[string]*Subscription{
			TopicProcessCreated: make(map[string]*Subscription),
			TopicProcessRemoved: make(map[string]*Subscription),
		},
	}
}

// AddSource dynamically adds a source to the broker.
func (b *Broker) AddSource(source <-chan []ProcessState) {
	// We could multiplex this instead of spawning multiple Goroutines, but it
	// is not expected that there will be many sources. If this assumption
	// changes we can refactor.
	go b.watchSource(source)
}

// watchSource listens for updates from a source and forwards them to all
// subscriptions.
func (b *Broker) watchSource(source <-chan []ProcessState) {
	for changes := range source {
		b.subsLock.Lock()

		created := b.subs[TopicProcessCreated]
		removed := b.subs[TopicProcessRemoved]

		for _, change := range changes {
			switch change.State {
			case StateCreated:
				for _, sub := range created {
					sub.handle(change.Process)
				}
			case StateRemoved:
				for _, sub := range removed {
					sub.handle(change.Process)
				}
			default:
				b.logger.Debug("ignoring state change", "change", change)
			}
		}
		b.subsLock.Unlock()
	}
}

// Subscribe adds a subscription to the broker.
func (b *Broker) Subscribe(topic Topic, matcher func(Process) bool) *Subscription {
	switch topic {
	case TopicProcessCreated, TopicProcessRemoved:
	// Valid.
	default:
		b.logger.Error("invalid topic", "topic", topic)
		return &Subscription{
			logger:      b.logger,
			unsubscribe: func() {},
			matcher:     func(Process) bool { return false },
		}
	}

	id := uuid.NewString()
	sub := &Subscription{
		logger:      b.logger,
		updates:     make(chan Process),
		unsubscribe: func() { b.unsubscribe(topic, id) },
		matcher:     matcher,
	}

	b.subsLock.Lock()
	// Note we don't check if id already exists based on the astronmically
	// small chance of collision, but we could change that if we want to be
	// even more safe.
	b.subs[topic][id] = sub
	b.subsLock.Unlock()

	return sub
}

// unsubscribe remores a subscription from the broker.
func (b *Broker) unsubscribe(topic Topic, id string) {
	b.subsLock.Lock()
	delete(b.subs[topic], id)
	b.subsLock.Unlock()
}

// Subscription tracks a subscription and allows unsubscribing.
type Subscription struct {
	logger *slog.Logger

	updates     chan Process
	unsubscribe func()
	matcher     func(Process) bool
}

// Updates returns the channel all updates for the subscription are sent on.
func (s *Subscription) Updates() <-chan Process { return s.updates }

// Unsubscribe removes the Subscription from the broker and stops its updates.
func (s *Subscription) Unsubscribe() {
	s.unsubscribe()
	close(s.updates)
}

// handle evaluates and handles the process. It will determine if the process
// matches the subscription criteria and sends it along to the updates channel.
func (s *Subscription) handle(process Process) {
	if s.matcher(process) {
		// Send the update to the subscription.
		select {
		case s.updates <- process:
		default:
			s.logger.Error("dropped process: slow receiver", "process", process)
		}
	}
}

// ProcessState represents the a state of a process.
type ProcessState struct {
	State   State
	Process Process
}

// State is the state of a process.
type State int

const (
	StateUnknown State = iota
	StateCreated
	StateRemoved
)
