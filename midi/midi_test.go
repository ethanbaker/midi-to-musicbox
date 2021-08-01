package midi_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/ethanbaker/midi-to-musicbox/midi"
)

func Test_MidiFile(t *testing.T) {
	/* TODO have Parse take no path, takes string of content in midi file
	// Read the raw contents of the MIDI file and set the contents equal
	// to file.Content
	f, err := os.Open("./testing/midi.mid")
	if err != nil {
		t.Log(err)
	}
	defer f.Close()
	*/

	// Parse the MIDI file
	var file midi.MidiFile
	file.Parse("./testing/midi.mid")

	// Save the output of the parse to a JSON file in the testing directory
	jsonString, err := json.Marshal(file)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("./testing/output.json", jsonString, os.ModePerm)
}
