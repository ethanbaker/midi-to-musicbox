package midi_test

import (
	"testing"

	"github.com/ethanbaker/midi-graphic"
)

func Test_ConvertImage(t *testing.T) {
	var f midi.MidiFile

	f.Parse("./testing/midi.mid")

	midi.CreateImage(f, "./testing/image.png")
}
