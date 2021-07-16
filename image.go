package midi

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// Midi note conversion to piano note
var midiToPiano = map[byte]int{
	21:  1,
	22:  2,
	23:  3,
	24:  4,
	25:  5,
	26:  6,
	27:  7,
	28:  8,
	29:  9,
	30:  10,
	31:  11,
	32:  12,
	33:  13,
	34:  14,
	35:  15,
	36:  16,
	37:  17,
	38:  18,
	39:  19,
	40:  20,
	41:  21,
	42:  22,
	43:  23,
	44:  24,
	45:  25,
	46:  26,
	47:  27,
	48:  28,
	49:  29,
	50:  30,
	51:  31,
	52:  32,
	53:  33,
	54:  34,
	55:  35,
	56:  36,
	57:  37,
	58:  38,
	59:  39,
	60:  40,
	61:  41,
	62:  42,
	63:  43,
	64:  44,
	65:  45,
	66:  46,
	67:  47,
	68:  48,
	69:  49,
	70:  50,
	71:  51,
	72:  52,
	73:  53,
	74:  54,
	75:  55,
	76:  56,
	77:  57,
	78:  58,
	79:  59,
	80:  60,
	81:  61,
	82:  62,
	83:  63,
	84:  64,
	85:  65,
	86:  66,
	87:  67,
	88:  68,
	89:  69,
	90:  70,
	91:  71,
	92:  72,
	93:  73,
	94:  74,
	95:  75,
	96:  76,
	97:  77,
	98:  78,
	99:  79,
	100: 80,
	101: 81,
	102: 82,
	103: 83,
	104: 84,
	105: 85,
	106: 86,
	107: 87,
	108: 88,
}

// Midi Note to name conversion
var midiToName = map[byte]string{
	21:  "A0",
	22:  "A#0/Bb0",
	23:  "B0",
	24:  "C1",
	25:  "C#1/Db1",
	26:  "D1",
	27:  "D#1/Eb1",
	28:  "E1",
	29:  "F1",
	30:  "F#1/Gb1",
	31:  "G1",
	32:  "G#1/Ab1",
	33:  "A1",
	34:  "A#1/Bb1",
	35:  "B1",
	36:  "C2",
	37:  "C#2/Db2",
	38:  "D2",
	39:  "D#2/Eb2",
	40:  "E2",
	41:  "F2",
	42:  "F#2/Gb2",
	43:  "G2",
	44:  "G#2/Ab2",
	45:  "A2",
	46:  "A#2/Bb2",
	47:  "B2",
	48:  "C3",
	49:  "C#3/Db3",
	50:  "D3",
	51:  "D#3/Eb3",
	52:  "E3",
	53:  "F3",
	54:  "F#3/Gb3",
	55:  "G3",
	56:  "G#3/Ab3",
	57:  "A3",
	58:  "A#3/Bb3",
	59:  "B3",
	60:  "C4",
	61:  "C#4/Db4",
	62:  "D4",
	63:  "D#4/Eb4",
	64:  "E4",
	65:  "F4",
	66:  "F#4/Gb4",
	67:  "G4",
	68:  "G#4/Ab4",
	69:  "A4",
	70:  "A#4/Bb4",
	71:  "B4",
	72:  "C5",
	73:  "C#5/Db5",
	74:  "D5",
	75:  "D#5/Eb5",
	76:  "E5",
	77:  "F5",
	78:  "F#5/Gb5",
	79:  "G5",
	80:  "G#5/Ab5",
	81:  "A5",
	82:  "A#5/Bb5",
	83:  "B5",
	84:  "C6",
	85:  "C#6/Db6",
	86:  "D6",
	87:  "D#6/Eb6",
	88:  "E6",
	89:  "F6",
	90:  "F#6/Gb6",
	91:  "G6",
	92:  "G#6/Ab6",
	93:  "A6",
	94:  "A#6/Bb6",
	95:  "B6",
	96:  "C7",
	97:  "C#7/Db7",
	98:  "D7",
	99:  "D#7/Db7",
	100: "E7",
	101: "F7",
	102: "F#7/Gb7",
	103: "G7",
	104: "G#7/Ab7",
	105: "A7",
	106: "A#7/Bb7",
	107: "B7",
	108: "C8",
}

type NoteTrack struct {
	name  string
	notes []float64
}

// Conversion rate from pixels to millimeters
const MILLI_CONVERSION_RATE = 0.2645833333

// CreateImage function creates an image based on the MidiFile
func CreateImage(file MidiFile) {
	// Specific parameters for music box
	var notes []string = []string{"C4", "D4", "E4", "F4", "G4"}
	USER_NOTE_SIZE := 5.0
	TRACK := 1

	// Important calculations for sheet size
	noteSize := int(USER_NOTE_SIZE/MILLI_CONVERSION_RATE) + 1
	height := noteSize * len(notes)
	width := 200

	// Get the top left and bottom right points of the image
	topLeft := image.Point{0, 0}
	bottomRight := image.Point{width, height}

	// Create an image
	img := image.NewRGBA(image.Rectangle{topLeft, bottomRight})

	// Create a color
	black := color.RGBA{0, 0, 0, 0xFF}
	white := color.RGBA{255, 255, 255, 0xFF}

	// Create a grid
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if y%noteSize == 0 {
				img.Set(x, y, black)
			} else {
				img.Set(x, y, white)
			}
		}
	}

	// Add the notes
	var noteTracks []NoteTrack
	for i := 0; i < len(notes); i++ {
		var track NoteTrack
		track.name = notes[i]
		noteTracks = append(noteTracks, track)
	}

	secondsPerTick := 60000.0 / float64(file.Tempo*int32(file.TimeDivision))

	for _, midiNote := range file.Tracks[TRACK].Notes {
		note := secondsPerTick * float64(midiNote.StartTime)

		for _, track := range noteTracks {
			if track.name == midiToName[midiNote.Key] {
				track.notes = append(track.notes, note)
				fmt.Println(len(track.notes))
				break
			}
		}
	}

	// Encode as PNG
	f, _ := os.Create("image.png")
	png.Encode(f, img)
}
