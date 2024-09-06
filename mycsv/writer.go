package mycsv

import (
	"io"
	"strings"
)

const MAX_PACKET_SIZE = 8000

type Writer struct {
	w      io.Writer
	buffer []byte
	cursor int
	err    error
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		buffer: make([]byte, MAX_PACKET_SIZE),
		w:      w,
	}
}

func (w *Writer) Write(record []string) {
	message := strings.Join(record, ",")

	w.write([]byte(message))
	w.write([]byte{'\n'})
}

// Writes data to buffer, or to inner writer if buffer is full
func (w *Writer) write(data []byte) {
	if w.err != nil {
		return
	}

	// Append to buffer all we can
	leftover := w.append(data)

	// While we have data left to write, flush and repeat
	for len(leftover) > 0 {
		w.Flush()

		leftover = w.append(leftover)
	}
}

// Writes all buffered data to inner writter
func (w *Writer) Flush() error {
	if w.err != nil {
		return w.err
	}

	// while we have data
	for w.cursor > 0 {

		// write all we can
		wrote, err := w.w.Write(w.buffer[:w.cursor])

		// update buffer and cursor
		w.discard(wrote)

		if err != nil {
			w.err = err
			return w.err
		}
	}

	return nil
}

// Discards `number` front bytes from buffer
func (w *Writer) discard(number int) {
	copy(w.buffer, w.buffer[number:w.cursor])
	w.cursor -= number
}

// Appends `data` to buffer and returns leftover (uncopied) data
func (w *Writer) append(data []byte) []byte {
	copied := copy(w.buffer[w.cursor:], data)
	w.cursor += copied

	return data[copied:]
}
