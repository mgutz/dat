package dat

import (
	"log"
	"os"
)

// EventReceiver gets events from dbr methods for logging purposes
type EventReceiver interface {
	Event(eventName string)
	EventKv(eventName string, kvs map[string]string)
	EventErr(eventName string, err error) error
	EventErrKv(eventName string, err error, kvs map[string]string) error
	Timing(eventName string, nanoseconds int64)
	TimingKv(eventName string, nanoseconds int64, kvs map[string]string)
}

type kvs map[string]string

// NullEventReceiver is a sentinel EventReceiver; use it if the caller doesn't supply one
type NullEventReceiver struct{}

var nullReceiver = &NullEventReceiver{}

// Event receives a simple notification when various events occur
func (n *NullEventReceiver) Event(eventName string) {
	// noop
}

// EventKv receives a notification when various events occur along with
// optional key/value data
func (n *NullEventReceiver) EventKv(eventName string, kvs map[string]string) {
	// noop
}

// EventErr receives a notification of an error if one occurs
func (n *NullEventReceiver) EventErr(eventName string, err error) error {
	return err
}

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data
func (n *NullEventReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	return err
}

// Timing receives the time an event took to happen
func (n *NullEventReceiver) Timing(eventName string, nanoseconds int64) {
	// noop
}

// TimingKv receives the time an event took to happen along with optional key/value data
func (n *NullEventReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	// noop
}

// LogEventReceiver is a sentinel EventReceiver; use it if the caller doesn't supply one
type LogEventReceiver struct {
	log *log.Logger
}

// NewLogEventReceiver creates a new LogEventReceiver.
func NewLogEventReceiver(prefix string) *LogEventReceiver {
	return &LogEventReceiver{
		log: log.New(os.Stdout, prefix, 0),
	}
}

var logReceiver = &LogEventReceiver{
	log: log.New(os.Stdout, "[dbr] ", 0),
}

// Event receives a simple notification when various events occur
func (ler *LogEventReceiver) Event(eventName string) {
	ler.log.Println(eventName)
}

// EventKv receives a notification when various events occur along with
// optional key/value data
func (ler *LogEventReceiver) EventKv(eventName string, kvs map[string]string) {
	ler.log.Println(eventName, kvs)
	// noop
}

// EventErr receives a notification of an error if one occurs
func (ler *LogEventReceiver) EventErr(eventName string, err error) error {
	ler.log.Println(eventName, err)
	return err
}

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data
func (ler *LogEventReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	ler.log.Println(eventName, err, kvs)
	return err
}

// Timing receives the time an event took to happen
func (ler *LogEventReceiver) Timing(eventName string, nanoseconds int64) {
	ler.log.Println(eventName, nanoseconds)
	// noop
}

// TimingKv receives the time an event took to happen along with optional key/value data
func (ler *LogEventReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	ler.log.Println(eventName, nanoseconds, kvs)
	// noop
}
