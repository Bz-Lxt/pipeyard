package pipeyard_test

import (
	"fmt"
	"testing"

	"github.com/Bz-Lxt/pipeyard/event"
)

func TestBusKeepsNewest(t *testing.T) {
	bus := event.NewBus()
	var last event.Event
	for i := 0; i < 10; i++ {
		last = event.Event{Kind: event.KindSubmit, Job: fmt.Sprintf("job-%02d", i), Note: "submit"}
		bus.Publish(last)
	}
	recent := bus.Recent(50)
	found := false
	for _, e := range recent {
		if e.Job == last.Job && e.Kind == event.KindSubmit {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Recent missing last submit %q among %d events", last.Job, len(recent))
	}
	kept := event.FilterKind(recent, event.KindSubmit)
	if len(kept) == 0 {
		t.Fatal("FilterKind dropped submit events")
	}
}
