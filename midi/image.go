package midi

import (
	"image"
	"math"
)

// CreateImage function creates an image based on the midiFile
func (img *MidiImage) CreateImage() {
	// Get information from the selected track in the file
	midiTrack := img.File.Tracks[img.Track]
	notes := midiTrack.Notes

	// Get the size of one channel (and the offset for every set amount of channels)
	channelSize := int((img.Height / MILLI_TO_PIXELS) / float64(len(img.Notes)))
	rawOffset := math.Abs(img.Height/MILLI_TO_PIXELS-float64(channelSize*len(img.Notes))) / float64(len(img.Notes))
	offset := int(rawOffset) + 1
	offsetCount := int(img.Height/MILLI_TO_PIXELS-float64(channelSize*len(img.Notes))) / offset
	offsetIndex := len(img.Notes)

	// Get the width and height of the sheet
	width := int(notes[len(notes)-1].StartTime/TIME_PER_COLUMN)*img.Speed + img.Speed
	height := int(img.Height/MILLI_TO_PIXELS) + 1

	// Create a new image using the bottom
	topLeft := image.Point{0, 0}
	bottomRight := image.Point{width, height}
	img.Picture = image.NewRGBA(image.Rectangle{topLeft, bottomRight})

	// Create a blank grid and keep track of when the channels start
	var channels []int
	for y := 0; y < height; y++ {
		tempOffsetIndex := 0
		if offsetIndex < offsetCount {
			tempOffsetIndex = 1
		}

		for x := 0; x < width; x++ {
			if y%(channelSize+offset*tempOffsetIndex) == 0 {
				img.Picture.Set(x, y+offset*tempOffsetIndex, black)
				if x == 0 {
					channels = append(channels, y+offset*tempOffsetIndex)
				}

				if offsetIndex == 0 {
					offsetIndex = offsetCount
				}
				offsetIndex--
			} else {
				img.Picture.Set(x, y+offset*tempOffsetIndex, white)
			}
		}
	}

	// Initialize the tracks to hold the notes
	var tracks []*noteTrack
	for i := 0; i < len(img.Notes); i++ {
		var track noteTrack
		track.name = img.Notes[i]
		tracks = append(tracks, &track)
	}

	// Add the midi notes from the file to the tracks
	for _, midiNote := range notes {
		// Get the note from the file
		note := float64(midiNote.StartTime) / TIME_PER_COLUMN

		for _, track := range tracks {
			if track.name == midiToNote[midiNote.Key] {
				track.notes = append(track.notes, note)
				break
			}
		}
	}

	// Draw the notes to the image
	for channel, track := range tracks {
		for _, note := range track.notes {
			x := img.Speed/2 + int(note*float64(img.Speed))
			y := channelSize/2 + channels[channel]
			r := (channelSize-4)/2 + 1
			drawCircle(img.Picture, x, y, r)
		}
	}

}

// TODO fix drawCircle function
// Helper function to draw a circle on the image
func drawCircle(img *image.RGBA, x0 int, y0 int, r int) {
	if r == 0 {
		img.Set(x0, y0, black)
		return
	}

	x, y := r-1, 0
	dy, dx := 1, 1
	err := dx - (r * 2)

	for x > y {
		img.Set(x0+x, y0+y, black)
		img.Set(x0+y, y0+x, black)
		img.Set(x0-y, y0+x, black)
		img.Set(x0-x, y0+y, black)
		img.Set(x0-x, y0-y, black)
		img.Set(x0-y, y0-x, black)
		img.Set(x0+y, y0-x, black)
		img.Set(x0+x, y0-y, black)

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
		}
	}

	img.Set(x0+(r/2)+1, y0+(r/2)+1, black)
	img.Set(x0-(r/2)-1, y0+(r/2)+1, black)
	img.Set(x0+(r/2)+1, y0-(r/2)-1, black)
	img.Set(x0-(r/2)-1, y0-(r/2)-1, black)

	drawCircle(img, x0, y0, r-1)
}
