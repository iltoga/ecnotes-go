package observer

import (
	"reflect"
	"testing"
)

var mockOnNotify OnNotify

type (
	mockData struct{}
	mockArgs struct{}
)

func TestNewObserver(t *testing.T) {
	tests := []struct {
		name string
		want Observer
	}{
		{
			name: "TestNewObserver",
			want: &ObserverImpl{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewObserver(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewObserver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObserver_AddListener(t *testing.T) {
	mockOnNotify = func(note interface{}, args ...interface{}) {}

	type fields struct {
		Listeners map[Event][]Listener
	}
	type args struct {
		event    Event
		listener Listener
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "TestObserver_AddListener:new",
			fields: fields{
				Listeners: nil,
			},
		},
		{
			name: "TestObserver_AddListener:append",
			fields: fields{
				Listeners: map[Event][]Listener{
					EVENT_UPDATE_NOTE_TITLES: {
						{
							OnNotify: mockOnNotify,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &ObserverImpl{
				listeners: tt.fields.Listeners,
			}
			o.AddListener(tt.args.event, tt.args.listener)
		})
	}
}

func TestObserver_Remove(t *testing.T) {
	type fields struct {
		Listeners map[Event][]Listener
	}
	type args struct {
		event Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "TestObserver_Remove",
			fields: fields{
				Listeners: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &ObserverImpl{
				listeners: tt.fields.Listeners,
			}
			o.Remove(tt.args.event)
		})
	}
}

func TestObserver_Notify(t *testing.T) {
	type fields struct {
		Listeners map[Event][]Listener
	}
	type args struct {
		event Event
		data  interface{}
		args  interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "TestObserver_Notify",
			fields: fields{
				Listeners: map[Event][]Listener{
					EVENT_UPDATE_NOTE_TITLES: {
						{
							OnNotify: mockOnNotify,
						},
					},
				},
			},
			args: args{
				event: EVENT_UPDATE_NOTE_TITLES,
				data:  mockData{},
				args:  mockArgs{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &ObserverImpl{
				listeners: tt.fields.Listeners,
			}
			o.Notify(tt.args.event, tt.args.data, tt.args.args)
		})
	}
}
