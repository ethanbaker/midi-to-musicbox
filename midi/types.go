package midi

import (
	"bufio"
	"image"
)

// Image Creator types ------------------------------------------------

// noteTrack type contains all of the notes for a key of a specific name
type noteTrack struct {
	name  string
	notes []float64
}

// MidiImage type is used to hold information for an image created from
// a MIDI file
type MidiImage struct {
	File    MidiFile `json:"file"`
	Track   int      `json:"track"`
	Speed   int      `json:"speed"`
	Height  float64  `json:"height"`
	Notes   []string `json:"notes"`
	Path    string   `json:"path"`
	Picture *image.RGBA
}

// Midi Parser types --------------------------------------------------

// midiEvent type is used to hold information from an event in a MIDI
// file
type midiEvent struct {
	Name      string `json:"name"`
	Key       byte   `json:"key"`
	Velocity  byte   `json:"velocity"`
	DeltaTick int32  `json:"deltaTick"`
}

// midiNote type is used to hold information from a note in a MIDI
// file
type midiNote struct {
	Key       byte  `json:"key"`
	Velocity  byte  `json:"velocity"`
	StartTime int32 `json:"startTime"`
	Duration  int32 `json:"duration"`
}

// midiTrack type is used to hold information from a track in a MIDI
// file
type midiTrack struct {
	Name       string      `json:"name"`
	Instrument string      `json:"instrument"`
	Min        byte        `json:"min"`
	Max        byte        `json:"max"`
	Events     []midiEvent `json:"events"`
	Notes      []midiNote  `json:"notes"`
}

// MidiFile type is used to hold information for a whole MIDI file
type MidiFile struct {
	Tracks       []midiTrack `json:"tracks"`
	Tempo        int32       `json:"tempo"`
	Bpm          int32       `json:"bpm"`
	TimeDivision int16       `json:"timeDivision"`
	Err          error       `json:"error"`

	Content string
	reader  *bufio.Reader
	atEof   bool
}
