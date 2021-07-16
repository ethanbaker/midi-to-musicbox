# Notes

* https://stackoverflow.com/questions/3660964/get-note-data-from-midi-file

## Header Chunk

* `data[0:4]`: Chunk id. Should be `MThd`, if not then error
* `data[4:8]`: Chunk size. Should be `6`, if not then error
* `data[8:10]`: Format type. Type `0` has one track with all events. Type `1`
  has two or more tracks. Type `2` is a combo of both tracks and is rarely
  used.
* `data[10:12]`: Number of tracks, between 1 and 65,535.
* `data[12:14]`: Time division. Used to decode the track event delta times into
  "real" time. Basically used as ticks per beat (or fps).

## Track Chunk

* `data[14:18]`: Chunk id. Should be `MTrk`, if not then error
* `data[18:22]`: Chunk size, varies depending on the amount of events in track.
* `data[22:]`: Track event data. Contains a stream of midi events.
