package observer

import "sync"

type (
	Event string

	OnNotify func(data interface{}, args ...interface{})

	Listener struct {
		OnNotify OnNotify
	}

	Observer interface {
		AddListener(event Event, listener Listener)
		Remove(event Event)
		Notify(event Event, data interface{}, args ...interface{})
	}

	ObserverImpl struct {
		listeners map[Event][]Listener
	}
)

var (
	observerInstance  *ObserverImpl
	singletonObserver sync.Once
)

// NewObserver get single global instance of Observer
func NewObserver() *ObserverImpl {
	singletonObserver.Do(func() {
		observerInstance = &ObserverImpl{}
	})
	return observerInstance
}

// AddListener add new listener in observer
func (o *ObserverImpl) AddListener(event Event, listener Listener) {
	if o.listeners == nil {
		o.listeners = map[Event][]Listener{}
	}
	o.listeners[event] = append(o.listeners[event], listener)
}

// Remove remove registered listener in observer
func (o *ObserverImpl) Remove(event Event) {
	delete(o.listeners, event)
}

// Notify send data & arg to registered listener based on event
func (o *ObserverImpl) Notify(event Event, data interface{}, args ...interface{}) {
	listeners, ok := o.listeners[event]
	if !ok {
		return
	}
	for _, listener := range listeners {
		go listener.OnNotify(data, args...)
	}
}
