package zapgray

import (
	"github.com/liqifyl/chat-go/internal/log/gelf"
	"os"

	"go.uber.org/zap"


	"go.uber.org/zap/zapcore"
)

// GelfCore implements the https://godoc.org/go.uber.org/zap/zapcore#Core interface
// Messages are written to a graylog endpoint using the GELF format + protocol
type GelfCore struct {
	g       *gelf.Gelf
	encoder zapcore.Encoder
	lv      zap.AtomicLevel
}

// NewGelfCore creates a new GelfCore with empty context.
func NewGelfCore(g *gelf.Gelf, lv zap.AtomicLevel) zapcore.Core {

	hostname, _ := os.Hostname()

	encoder := zapcore.NewJSONEncoder(NewGraylogEncoderConfig())

	encoder.AddString("version", "1.1")
	encoder.AddString("host", hostname)

	return &GelfCore{
		g:       g,
		encoder: encoder,
		lv:      lv,
	}
}

// Write writes messages to the configured Graylog endpoint.
func (gc *GelfCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {

	ff := getGrayFields(fields)

	// Encode the zap fields from fields to JSON with proper types.
	buf, err := gc.encoder.EncodeEntry(entry, ff)
	if err != nil {
		return err
	}

	gc.g.Log(buf.Bytes())
	return nil
}

// With adds structured context to the logger.
func (gc *GelfCore) With(fields []zapcore.Field) zapcore.Core {
	clone := gc.clone()
	addFields(clone.encoder, fields)
	return clone
}

// Sync is a no-op.
func (gc *GelfCore) Sync() error {
	return nil
}

// Check determines whether the supplied entry should be logged.
func (gc *GelfCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if gc.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, gc)
	}
	return checkedEntry
}

// Enabled only enables info messages and above.
func (gc *GelfCore) Enabled(level zapcore.Level) bool {
	return gc.lv.Enabled(level)
}

func (gc *GelfCore) clone() *GelfCore {
	return &GelfCore{
		g:       gc.g,
		encoder: gc.encoder.Clone(),
	}
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {

	ff := getGrayFields(fields)
	for i := range ff {
		ff[i].AddTo(enc)
	}
}

func getGrayFields(fields []zapcore.Field) []zapcore.Field {

	ret := make([]zapcore.Field, 0)

	for i := range fields {

		f := fields[i]

		if f.Key != "full_message" {
			f.Key = "_udef-" + f.Key
		}

		ret = append(ret, f)

	}

	return ret
}