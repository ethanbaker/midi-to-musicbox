package midi_test

import (
	"image/png"
	"log"
	"os"
	"testing"

	"github.com/ethanbaker/midi-to-musicbox/midi"
)

func Test_ConvertImage(t *testing.T) {
	var img midi.MidiImage

	img.File.Parse("./testing/midi.mid")
	if img.File.Err != nil {
		log.Fatal(img.File.Err)
	}

	img.Notes = []string{"C4", "C#4/Db4", "D4", "D#4/Eb4", "E4", "F4", "F#4/Gb4", "G4", "G#4/Ab4", "A4", "A#4/Bb4", "B4", "C5", "C#5/Db5", "D5", "D#5/Eb5", "E5", "F5", "F#5", "F#5/Gb5", "G5", "G#5/Ab5", "A5", "A#5/Bb5", "B5"}
	img.Track = 1
	img.Speed = 200
	img.Height = float64(len(img.Notes) * 5)

	// Create the image
	img.CreateImage()

	// Encode the image as a PNG
	f, err := os.Create("./testing/image.png")
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(f, img.Picture)
}

/*
file: the midi file holding the notes
noteLabels: the actual notes present on the music box (in order)
speed: pixels per second, spacing out notes
height: height of music box sheet in mm
outputPath: path that file gets saved to
*/
