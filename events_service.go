package apid

type EventSelector string

type Event interface{}

type EventHandler interface {
	Handle(event Event)
}

type EventHandlerFunc func(event Event)

type EventsService interface {
	/* publish an event to the selector.
	It will send a copy of the delivered event to the returned channel, after all listeners have responded to the event.
	Call "Emit()" for non-blocking, "<-Emit()" for blocking.
	*/
	Emit(selector EventSelector, event Event) chan Event

	// publish an event to the selector, call the passed handler when all listeners have responded to the event
	EmitWithCallback(selector EventSelector, event Event, handler EventHandlerFunc)

	// when an event matching selector occurs, run the provided handler
	Listen(selector EventSelector, handler EventHandler)

	// when an event matching selector occurs, run the provided handler function
	ListenFunc(selector EventSelector, handler EventHandlerFunc)

	// when an event matching selector occurs, run the provided handler function and stop listening
	ListenOnceFunc(selector EventSelector, handler EventHandlerFunc)

	// remove a listener
	StopListening(selector EventSelector, handler EventHandler)

	// shut it down
	Close()
}

const EventDeliveredSelector EventSelector = "event delivered"

type EventDeliveryEvent struct {
	Description string
	Selector    EventSelector
	Event       Event
	Count       int
}

// use reflect.DeepEqual to compare this type
type PluginsInitializedEvent struct {
	Description string
	// using slice member will make the type "PluginsInitializedEvent" uncomparable
	Plugins []PluginData
}

type PluginData struct {
	Name      string
	Version   string
	ExtraData map[string]interface{}
}
