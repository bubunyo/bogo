package bogo

import (
	"encoding/binary"
	"fmt"
	"time"
)

func encodeTimestamp(timestamp int64) ([]byte, error) {
	buf := make([]byte, 9) // 1 byte type + 8 bytes int64
	buf[0] = byte(TypeTimestamp)
	binary.LittleEndian.PutUint64(buf[1:], uint64(timestamp))
	return buf, nil
}

func decodeTimestamp(data []byte) (int64, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("timestamp decode error: insufficient data, need 8 bytes, got %d", len(data))
	}

	timestamp := int64(binary.LittleEndian.Uint64(data[:8]))
	return timestamp, nil
}

// Helper function to encode time.Time as timestamp
func encodeTimeValue(t time.Time) ([]byte, error) {
	return encodeTimestamp(t.UnixMilli())
}
