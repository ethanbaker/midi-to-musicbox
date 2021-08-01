package midi

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

// Helper functions ---------------------------------------------------

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

// MidiFile functions -------------------------------------------------

// handleError handles all errors and ignores any EOF errors
func (f *MidiFile) handleError(err error) {
	if err == io.EOF {
		return
	}

	f.Err = err
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
	if val > 0x7F {
		// Extract the bottom 7 bits of the read byte
		val &= 0x7F

		// Keep reading bytes until the compression has stopped
		b, err = f.reader.ReadByte()
		if err != nil {
			f.handleError(err)
		}
		for b > 0x7F {
			// Read the next byte
			b, err = f.reader.ReadByte()
			if err != nil {
				f.handleError(err)
			}

			// Add the next byte to the value
			val = (val << 7) | int32(b&0x7F)
		}
	}

	return val
}

// Parse parses the MIDI file and creates a struct to hold the info
func (f *MidiFile) Parse(inputPath string) {
	// Open the MIDI file as a stream
	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a scanner to read all of the bytes
	f.reader = bufio.NewReader(file)
	var b []byte

	// Read the File ID
	b, err = f.reader.Peek(4)
	if err != nil {
		f.handleError(err)
	}
	fileId := string(b)
	if fileId != "MThd" {
		f.handleError(fmt.Errorf("Corrupted MIDI header"))
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
		f.handleError(fmt.Errorf("Corrupted MIDI header"))
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
	_ = byteToInt16(b)
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

	// Read the track chunks
	for trackIndex := 0; trackIndex < int(trackNumber); trackIndex++ {
		// Add the track to the list of tracks
		var track midiTrack
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
			f.handleError(fmt.Errorf("Track ID of track " + fmt.Sprint(trackIndex) + " is not `MTrk`, aborting!"))
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
		_ = byteToInt32(b)
		_, err = f.reader.Discard(4)
		if err != nil {
			f.handleError(err)
		}

		// Initalize the previous status byte
		var previousStatus byte

		// Make sure we catch when we are at the end of the track or EOF
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

				// Create a new midiEvent and add it to the current track
				event := midiEvent{"NoteOff", noteId, noteVelocity, statusTimeDelta}
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

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

				// Create a new midiEvent and add it to the current track
				var event midiEvent
				if noteVelocity == 0 {
					event = midiEvent{"NoteOff", noteId, noteVelocity, statusTimeDelta}
				} else {
					event = midiEvent{"NoteOn", noteId, noteVelocity, statusTimeDelta}
				}
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

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

				// Create a new midiEvent and add it to the current track
				var event midiEvent
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

				// Create a new midiEvent and add it to the current track
				var event midiEvent
				event.Name = "Other"
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

			case VoiceProgramChange:
				previousStatus = status

				// Get the program id
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new midiEvent and add it to the current track
				var event midiEvent
				event.Name = "Other"
				f.Tracks[trackIndex].Events = append(f.Tracks[trackIndex].Events, event)

			case VoiceChannelPressure:
				previousStatus = status

				// Get the channel pressure
				_, err = f.reader.ReadByte()
				if err != nil {
					f.handleError(err)
				}

				// Create a new midiEvent and add it to the current track
				var event midiEvent
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

				// Create a new midiEvent and add it to the current track
				var event midiEvent
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
						_, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

					case MetaText:
						_ = f.readString(length)

					case MetaCopyright:
						_ = f.readString(length)

					case MetaTrackName:
						f.Tracks[trackIndex].Name = f.readString(length)

					case MetaInstrumentName:
						f.Tracks[trackIndex].Instrument = f.readString(length)

					case MetaLyrics:
						_ = f.readString(length)

					case MetaMarker:
						_ = f.readString(length)

					case MetaCuePoint:
						_ = f.readString(length)

					case MetaChannelPrefix:
						_ = f.readString(length)

					case MetaEndOfTrack:
						_ = f.readString(length)
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
							f.Bpm = (60000000 / f.Tempo)
						}

					case MetaSMPTEOffset:
						// Get various offset attributes
						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

					case MetaTimeSignature:
						// Get various time signature attributes
						_, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

					case MetaKeySignature:
						// Get key attributes
						_, err := f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

						_, err = f.reader.ReadByte()
						if err != nil {
							f.handleError(err)
						}

					case MetaSequencerSpecific:
						_ = f.readString(length)

					default:
						fmt.Println("Warning! Unrecognized MetaEvent " + fmt.Sprint(nType))
					}
				}

				if status == 0xF0 {
					_ = f.readString(f.readValue())
				} else if status == 0xF7 {
					_ = f.readString(f.readValue())
				}

			default:
				fmt.Println("Warning! Unrecognized status byte: " + fmt.Sprint(status))

			}
		}
	}

	// Convert time events to notes
	for index, _ := range f.Tracks {
		var notesBeingProcessed []midiNote
		var wallTime int32

		for _, event := range f.Tracks[index].Events {
			wallTime += event.DeltaTick

			if event.Name == "NoteOn" {
				// Add an 'NoteOn' to the processing notes
				note := midiNote{event.Key, event.Velocity, wallTime, 0}
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
}
