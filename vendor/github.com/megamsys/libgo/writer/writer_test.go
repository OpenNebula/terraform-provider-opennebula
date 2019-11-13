package writer

/*
import (
	"time"

	"github.com/megamsys/vertice/provision"
	"gopkg.in/check.v1"
)

type WriterSuite struct {
	//	conn *db.Storage
}

var _ = check.Suite(&WriterSuite{})

func (s *WriterSuite) TestLogWriter(c *check.C) {
	a := provision.Box{}
	writer := LogWriter{Box: &a}
	data := []byte("ble")
	_, err := writer.Write(data)
	c.Assert(err, check.IsNil)
}

func (s *WriterSuite) TestLogWriterShouldReturnTheDataSize(c *check.C) {
	a := provision.Box{}
	writer := LogWriter{Box: &a}
	data := []byte("ble")
	n, err := writer.Write(data)
	c.Assert(err, check.IsNil)
	c.Assert(n, check.Equals, len(data))
}

func (s *WriterSuite) TestLogWriterAsync(c *check.C) {
	a := provision.Box{}
	writer := LogWriter{Box: &a}
	writer.Async()
	data := []byte("ble")
	_, err := writer.Write(data)
	c.Assert(err, check.IsNil)
	writer.Close()
	err = writer.Wait(5 * time.Second)
	c.Assert(err, check.IsNil)
}

func (s *WriterSuite) TestLogWriterAsyncCopySlice(c *check.C) {
	a := provision.Box{}
	writer := LogWriter{Box: &a}
	writer.Async()
	for i := 0; i < 100; i++ {
		data := []byte("ble")
		_, err := writer.Write(data)
		data[0] = 'X'
		c.Assert(err, check.IsNil)
	}
	writer.Close()
	err := writer.Wait(5 * time.Second)
	c.Assert(err, check.IsNil)
}
*/
