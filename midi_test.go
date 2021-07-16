package midi_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/ethanbaker/midi-graphic"
)

func Test_MidiFile(t *testing.T) {
	var f midi.MidiFile

	f.Parse("./midi.mid", "./output.mid")

	jsonString, err := json.Marshal(f)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("output.json", jsonString, os.ModePerm)
}
