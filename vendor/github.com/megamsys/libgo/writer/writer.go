package writer

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"time"
)

type Logger interface {
	Log(string, string, string) error
}

type LogWriter struct {
	Box    Logger
	Source string
	msgCh  chan []byte
	doneCh chan bool
}

func NewLogWriter(l Logger) LogWriter {
	logWriter := LogWriter{Box: l}
	logWriter.Async()
	return logWriter
}

func (w *LogWriter) Async() {
	w.msgCh = make(chan []byte, 1000)
	w.doneCh = make(chan bool)
	go func() {
		defer close(w.doneCh)
		for msg := range w.msgCh {
			err := w.write(msg)
			if err != nil {
				log.Errorf("[LogWriter] failed to write async logs: %s", err)
				return
			}
		}
	}()
}

func (w *LogWriter) Close() {
	if w.msgCh != nil {
		close(w.msgCh)
	}
}

func (w *LogWriter) Wait(timeout time.Duration) error {
	if w.msgCh == nil {
		return nil
	}
	select {
	case <-w.doneCh:
	case <-time.After(timeout):
		return errors.New("timeout waiting for writer to finish")
	}
	return nil
}

// Write writes and logs the data.
func (w *LogWriter) Write(data []byte) (int, error) {
	if w.msgCh == nil {
		return len(data), w.write(data)
	}
	copied := make([]byte, len(data))
	copy(copied, data)
	w.msgCh <- copied
	return len(data), nil
}

func (w *LogWriter) write(data []byte) error {
	return w.Box.Log(string(data), "megd", "box")
}
