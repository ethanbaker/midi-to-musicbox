package midi

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

// Constants
const (
	// For MidiTracks
	Max_Note = 64
	Min_Note = 64

	// For events
	VoiceNoteOff         = 0x80
	VoiceNoteOn          = 0x90
	VoiceAftertouch      = 0xA0
	VoiceControlChange   = 0xB0
	VoiceProgramChange   = 0xC0
	VoiceChannelPressure = 0xD0
	VoicePitchBend       = 0xE0
	SystemExclusive      = 0xF0

	// For meta events
	MetaSequence          = 0x00
	MetaText              = 0x01
	MetaCopyright         = 0x02
	MetaTrackName         = 0x03
	MetaInstrumentName    = 0x04
	MetaLyrics            = 0x05
	MetaMarker            = 0x06
	MetaCuePoint          = 0x07
	MetaChannelPrefix     = 0x20
	MetaEndOfTrack        = 0x2F
	MetaSetTempo          = 0x51
	MetaSMPTEOffset       = 0x54
	MetaTimeSignature     = 0x58
	MetaKeySignature      = 0x59
	MetaSequencerSpecific = 0x7F
)

// MidiEvent type used to hold information from an event
type MidiEvent struct {
	Name      string `json:"name"`
	Key       byte   `json:"key"`
	Velocity  byte   `json:"velocity"`
	DeltaTick int32  `json:"deltaTick"`
}

// MidiNote type used to hold information from a note
type MidiNote struct {
	Key       byte  `json:"key"`
	Velocity  byte  `json:"velocity"`
	StartTime int32 `json:"startTime"`
	Duration  int32 `json:"duration"`
}

// MidiTrack type used to hold information from a track
type MidiTrack struct {
	Name       string      `json:"name"`
	Instrument string      `json:"instrument"`
	Min        byte        `json:"min"`
	Max        byte        `json:"max"`
	Events     []MidiEvent `json:"events"`
	Notes      []MidiNote  `json:"notes"`
}

type MidiFile struct {
	Tracks       []MidiTrack `json:"tracks"`
	Tempo        int32       `json:"tempo"`
	TimeDivision int16       `json:"timeDivision"`
	reader       *bufio.Reader
	atEof        bool
}

// Helper functions

// byteToInt32 converts an array of bytes to an 32 bit integer
func byteToInt32(b []byte) int32 {
	if len(b) != 4 {
		return -1
	}

	var n int32
	n |= int32(b[0])
	n |= int32(b[1])
	n |= int32(b[2])
	n |= int32(b[3])

	return n
}

// byteToInt16 converts an array of bytes to an 16 bit integer
func byteToInt16(b []byte) int16 {
	if len(b) != 2 {
		return -1
	}

	var n int16
	n |= int16(b[0])
	n |= int16(b[1])

	return n
}

// handleError handles all errors and checks specifically for an EOF error
func (f *MidiFile) handleError(err error) {
	if err == io.EOF {
		return
	}

	log.Fatal(err)
}

// readString reads 'n' bytes from the scanner
func (f *MidiFile) readString(n int32) string {
	s := ""

	for i := 0; i < int(n); i++ {
		c, err := f.reader.ReadByte()
		if err != nil {
			f.handleError(err)
		}
		s += string(c)
	}

	return s
}

// readValue reads a compressed MIDI value
func (f *MidiFile) readValue() int32 {
	var val int32
	var b byte

	// Read the first byte
	v, err := f.reader.ReadByte()
	if err != nil {
		f.handleError(err)
	}
	val = int32(v)

	// Check if more bytes need reading
	if val > 127 {
		// Extract the bottom 7 bits of the read byte
		val &= 127

		// Keep reading bytes until the compression has stopped
		b, err = f.reader.ReadByte()
		if err != nil {
			f.handleError(err)
		}
		for b > 127 {
			// Read the next byte
			b, err = f.reader.ReadByte()
			if err != nil {
				f.handleError(err)
			}

			// Add the next byte to the value
			val = (val << 7) | int32(b&127)
		}
	}

	return val
}

func (f *MidiFile) Parse(inputPath string) bool {
	// Open the MIDI file as a stream
	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a scanner to read all of the bytes
	f.reader = bufio.NewReader(file)

	// Filler variables to save memory
	var b []byte

	// Read the MIDI Header
	fmt.Println("Starting parse")

	// Read the File ID
	b, err = f.reader.Peek(4)
	if err != nil {
		f.handleError(err)
	}
	fileId := string(b)
	if fileId != "MThd" {
		log.Fatal("File ID is not 'MThd', aborting!")
	}
	_, err = f.reader.Discard(4)
	if err != nil {
		f.handleError(err)
	}

	// Read the header length
	b, err = f.reader.Peek(4)
	if err != nil {
		f.handleError(err)
	}
	headerLength := byteToInt32(b)
	if headerLength != 6 {
		log.Fatal("Header length is not '6', aborting!")
	}
	_, err = f.reader.Discard(4)
	if err != nil {
		f.handleError(err)
	}

	// Read the format type
	b, err = f.reader.Peek(2)
	if err != nil {
		f.handleError(err)
	}
	format := byteToInt16(b)
	_, err = f.reader.Discard(2)
	if err != nil {
		f.handleError(err)
	}

	// Read the number of tracks
	b, err = f.reader.Peek(2)
	if err != nil {
		f.handleError(err)
	}
	trackNumber := byteToInt16(b)
	_, err = f.reader.Discard(2)
	if err != nil {
		f.handleError(err)
	}

	// Read the time division
	b, err = f.reader.Peek(2)
	if err != nil {
		f.handleError(err)
	}
	f.TimeDivision = byteToInt16(b)
	_, err = f.reader.Discard(2)
	if err != nil {
		f.handleError(err)
	}

	fmt.Println("Parsed file id:", fileId)
	fmt.Println("Parsed header length:", headerLength)
	fmt.Println("Parsed format number:", format)
	fmt.Println("Parsed track number:", trackNumber)
	fmt.Println("Parsed time division:", f.TimeDivision)

	// Read the track chunks
	for trackIndex := 0; trackIndex < int(trackNumber); trackIndex++ {
		fmt.Println("========== Starting track", trackIndex)

		// Add the track to the list of tracks
		var track MidiTrack
		track.Min = 64
		track.Max = 64
		f.Tracks = append(f.Tracks, track)

		// Read the track header
		b, err = f.reader.Peek(4)
		if err != nil {
			f.handleError(err)
		}
		trackId := string(b)
		if trackId != "MTrk" {
			log.Fatal("Track ID of track " + fmt.Sprint(trackIndex) + " is not `MTrk`, aborting!")
		}
		_, err = f.reader.Discard(4)
		if err != nil {
			f.handleError(err)
		}

		// Read the track length
		b, err = f.reader.Peek(4)
		if err != nil {
			f.handleError(err)
		}
		trackLength := byteToInt32(b)
		_, err = f.reader.Discard(4)
		if err != nil {
			f.handleError(err)
		}

		fmt.Println("Parsed track ID:", trackId)
		fmt.Println("Parsed track length:", trackLength)

		// Read the rest of the track data
		var previousStatus byte

		endOfTrack := false
		f.atEof = false
		for !f.atEof && !endOfTrack {
			// Read the timecode from MIDI stream
			statusTimeDelta := f.readValue()

			// Read the first byte of the message, which may be the status byte
			b, err := f.reader.Peek(1)
			if err != nil {
				f.handleError(err)
			}
			status := b[0]

			// If the status byte was not set (omitted for compression, revert to the previous byte. Otherwise, progress forward
			if status < 0x80 {
				status = previousStatus
			} else {
				_, err := f.reader.Discard(1)
				if err != nil {
					f.handleError(err)
				}
			}

			// Read and parse different event types
			switch status & 0xF0 {
			case VoiceNoteOff:
				previousStatus = status

				// Get the note id
				noteId, err := f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Get the note velocity
				noteVelocity, err := f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new MidiEvent and add it to the current track
				event := MidiEvent{"NoteOff", noteId, noteVelocity, statusTimeDelta}
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

				fmt.Println("NoteOff added")

			case VoiceNoteOn:
				previousStatus = status

				// Get the note id
				noteId, err := f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Get the note velocity
				noteVelocity, err := f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new MidiEvent and add it to the current track
				var event MidiEvent
				if noteVelocity == 0 {
					event = MidiEvent{"NoteOff", noteId, noteVelocity, statusTimeDelta}
				} else {
					event = MidiEvent{"NoteOn", noteId, noteVelocity, statusTimeDelta}
				}
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

				fmt.Println("NoteOn added")

			case VoiceAftertouch:
				previousStatus = status

				// Get the note id
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Get the note velocity
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new MidiEvent and add it to the current track
				var event MidiEvent
				event.Name = "Other"
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

			case VoiceControlChange:
				previousStatus = status

				// Get the control id
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Get the control value
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new MidiEvent and add it to the current track
				var event MidiEvent
				event.Name = "Other"
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

			case VoiceProgramChange:
				previousStatus = status

				// Get the program id
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new MidiEvent and add it to the current track
				var event MidiEvent
				event.Name = "Other"
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

			case VoiceChannelPressure:
				previousStatus = status

				// Get the channel pressure
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new MidiEvent and add it to the current track
				var event MidiEvent
				event.Name = "Other"
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

			case VoicePitchBend:
				previousStatus = status

				// Get the LS7B
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Get the MS7B
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new MidiEvent and add it to the current track
				var event MidiEvent
				event.Name = "Other"
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

			case SystemExclusive:
				previousStatus = 0

				// If the event is a meta message
				if status == 0xFF {

					// Get the length and type of the event
					nType, err := f.reader.ReadByte()
					if err != nil {
						f.handleError(err)
					}
					length := f.readValue()

					switch nType {
					case MetaSequence:
						num1, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						num2, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}
						fmt.Println("Sequence number: " + fmt.Sprint(num1) + fmt.Sprint(num2))

					case MetaText:
						fmt.Println("Text: " + f.readString(length))

					case MetaCopyright:
						fmt.Println("Copyright: " + f.readString(length))

					case MetaTrackName:
						f.Tracks[trackIndex].Name = f.readString(length)
						fmt.Println("Track name: " + f.Tracks[trackIndex].Name)

					case MetaInstrumentName:
						f.Tracks[trackIndex].Instrument = f.readString(length)
						fmt.Println("Instrument name: " + f.Tracks[trackIndex].Instrument)

					case MetaLyrics:
						fmt.Println("Lyrics: " + f.readString(length))

					case MetaMarker:
						fmt.Println("Marker: " + f.readString(length))

					case MetaCuePoint:
						fmt.Println("Cue: " + f.readString(length))

					case MetaChannelPrefix:
						fmt.Println("Prefix: " + f.readString(length))

					case MetaEndOfTrack:
						fmt.Println("End of track")
						endOfTrack = true

					case MetaSetTempo:
						// Tempo is in microseconds per quarter note
						if f.Tempo == 0 {
							// Get the three values for the tempo
							t1, err := f.reader.ReadByte()
							if err != nil {
								f.handleError(err)
							}

							t2, err := f.reader.ReadByte()
							if err != nil {
								f.handleError(err)
							}

							t3, err := f.reader.ReadByte()
							if err != nil {
								f.handleError(err)
							}

							// Set the tempo
							f.Tempo |= int32(t1) << 16
							f.Tempo |= int32(t2) << 8
							f.Tempo |= int32(t3) << 0

							// Display the tempo (and bpm)
							bpm := (60000000 / f.Tempo)

							fmt.Println("Tempo: " + fmt.Sprint(f.Tempo) + " (BPM: " + fmt.Sprint(bpm) + ")")
						}

					case MetaSMPTEOffset:
						// Get the attributes
						h, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						m, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						s, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						fr, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						ff, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						// Display the attributes
						fmt.Println("SMPTE: H:" + fmt.Sprint(h) + " M:" + fmt.Sprint(m) + " S:" + fmt.Sprint(s) + " FR:" + fmt.Sprint(fr) + "FF:" + fmt.Sprint(ff))

					case MetaTimeSignature:
						// Get the attributes
						ts1, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						ts2, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						cpt, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						per24c, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						// Display the attributes
						fmt.Println("Time signature: " + fmt.Sprint(ts1) + " / " + fmt.Sprint(2<<ts2))
						fmt.Println("Clocks per tick: " + fmt.Sprint(cpt))
						fmt.Println("32 per 24 clocks: " + fmt.Sprint(per24c))

					case MetaKeySignature:
						// Get the attributes
						keySignature, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						minorKey, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						// Display the attributes
						fmt.Println("Key signature: " + fmt.Sprint(keySignature))
						fmt.Println("Minor key: " + fmt.Sprint(minorKey))

					case MetaSequencerSpecific:
						fmt.Println("Sequencer specifics: " + f.readString(length))

					default:
						fmt.Println("Warning! Unrecognized MetaEvent " + fmt.Sprint(nType))
					}
				}

				if status == 0xF0 {
					fmt.Println("System exclusive begin: " + f.readString(f.readValue()))
				} else if status == 0xF7 {
					fmt.Println("System exclusive end: " + f.readString(f.readValue()))
				}

			default:
				fmt.Println("Unrecognized status byte: " + fmt.Sprint(status))

			}
		}
	}

	// Convert time events to notes
	for index, _ := range f.Tracks {
		var notesBeingProcessed []MidiNote
		var wallTime int32

		for _, event := range f.Tracks[index].Events {
			wallTime += event.DeltaTick

			if event.Name == "NoteOn" {
				// Add an 'NoteOn' to the processing notes
				note := MidiNote{event.Key, event.Velocity, wallTime, 0}
				notesBeingProcessed = append(notesBeingProcessed, note)
			} else if event.Name == "NoteOff" {
				// Remove an 'NoteOn' if it exists from processing notes
				for i, note := range notesBeingProcessed {
					if event.Key == note.Key {
						// Set the note's duration and add it to the track
						note.Duration = wallTime - note.StartTime
						f.Tracks[index].Notes = append(f.Tracks[index].Notes, note)

						// Change the min/max note of the track
						if note.Key < f.Tracks[index].Min {
							f.Tracks[index].Min = note.Key
						}

						if note.Key > f.Tracks[index].Max {
							f.Tracks[index].Max = note.Key
						}

						if i < len(notesBeingProcessed) {
							notesBeingProcessed = append(notesBeingProcessed[:i], notesBeingProcessed[i+1:]...)
						}
					}
				}
			}
		}
	}

	return true
}
