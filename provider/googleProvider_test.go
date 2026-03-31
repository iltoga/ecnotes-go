package provider

import (
	"context"
	"testing"
	"time"

	"github.com/iltoga/ecnotes-go/model"
	"github.com/sirupsen/logrus"
)

func TestUpdateNoteNotifier_EnqueuesNote(t *testing.T) {
	logger := logrus.New()
	gp := &GoogleProvider{
		updateQueue: make(chan *model.Note, 10),
		deleteQueue: make(chan int, 10),
		logger:      logger,
	}

	listener := gp.UpdateNoteNotifier()

	note := &model.Note{
		ID:    123,
		Title: "Test Note",
	}

	// The listener expects the encrypted note as args[2] currently, according to implementation
	// listener.OnNotify(note, nil, nil, note)
	listener.OnNotify("dummy", nil, nil, note) 

	// Verify the note is in the queue
	select {
	case queuedNote := <-gp.updateQueue:
		if queuedNote.ID != 123 {
			t.Errorf("Expected note ID 123 in queue, got %v", queuedNote.ID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Note was not enqueued in time")
	}
}

func TestDeleteNoteNotifier_EnqueuesDelete(t *testing.T) {
	logger := logrus.New()
	gp := &GoogleProvider{
		deleteQueue: make(chan int, 10),
		logger:      logger,
	}

	listener := gp.DeleteNoteNotifier()
	note := &model.Note{ID: 456}

	listener.OnNotify(note)

	// Verify the delete ID is in the queue
	select {
	case queuedID := <-gp.deleteQueue:
		if queuedID != 456 {
			t.Errorf("Expected note ID 456 in delete queue, got %v", queuedID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Delete ID was not enqueued in time")
	}
}

func TestInitWorker_ContextCancellation(t *testing.T) {
	logger := logrus.New()
	gp := &GoogleProvider{
		updateQueue: make(chan *model.Note, 10),
		deleteQueue: make(chan int, 10),
		logger:      logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	// Start worker, then immediately cancel to make sure it doesn't panic and exits cleanly
	gp.InitWorker(ctx)
	cancel()
	
	// Small delay to allow goroutine to print shutdown log
	time.Sleep(100 * time.Millisecond)
	
	// If it hasn't crashed, test passes.
}
