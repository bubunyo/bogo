package bogo

import (
	"io"
)

// StreamEncoder writes bogo values to an output stream, similar to json.Encoder
type StreamEncoder struct {
	w       io.Writer
	encoder *Encoder
}

// NewEncoder creates a new StreamEncoder that writes to w, similar to json.NewEncoder
func NewEncoder(w io.Writer) *StreamEncoder {
	return &StreamEncoder{
		w:       w,
		encoder: NewConfigurableEncoder(),
	}
}

// NewEncoderWithOptions creates a StreamEncoder with custom configuration options
func NewEncoderWithOptions(w io.Writer, options ...EncoderOption) *StreamEncoder {
	return &StreamEncoder{
		w:       w,
		encoder: NewConfigurableEncoder(options...),
	}
}

// Encode encodes v and writes it to the stream, similar to json.Encoder.Encode
func (enc *StreamEncoder) Encode(v any) error {
	data, err := enc.encoder.Encode(v)
	if err != nil {
		return err
	}
	
	_, err = enc.w.Write(data)
	return err
}

// SetEncoder allows setting a custom encoder instance
func (enc *StreamEncoder) SetEncoder(e *Encoder) {
	if e != nil {
		enc.encoder = e
	}
}

// StreamDecoder reads and decodes bogo values from an input stream, similar to json.Decoder
type StreamDecoder struct {
	r       io.Reader
	decoder *Decoder
}

// NewDecoder creates a new StreamDecoder that reads from r, similar to json.NewDecoder
func NewDecoder(r io.Reader) *StreamDecoder {
	return &StreamDecoder{
		r:       r,
		decoder: NewConfigurableDecoder(),
	}
}

// NewDecoderWithOptions creates a StreamDecoder with custom configuration options
func NewDecoderWithOptions(r io.Reader, options ...DecoderOption) *StreamDecoder {
	return &StreamDecoder{
		r:       r,
		decoder: NewConfigurableDecoder(options...),
	}
}

// Decode reads the next bogo value from the stream and stores it in v, similar to json.Decoder.Decode
func (dec *StreamDecoder) Decode(v any) error {
	result, err := dec.decoder.DecodeFrom(dec.r)
	if err != nil {
		return err
	}
	
	return assignResult(result, v)
}

// SetDecoder allows setting a custom decoder instance
func (dec *StreamDecoder) SetDecoder(d *Decoder) {
	if d != nil {
		dec.decoder = d
	}
}