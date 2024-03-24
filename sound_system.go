package main

import (
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type DecodedSound struct {
	name     string
	streamer beep.StreamSeekCloser
	format   beep.Format
}

type SoundSystem struct {
	streamer beep.StreamSeekCloser
	format   beep.Format
	ctrl     *beep.Ctrl
}

func NewSoundSystem() *SoundSystem {
	return &SoundSystem{}
}

func (sound_system *SoundSystem) InitSpeaker(format beep.Format) {
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
}

func (sound_system *SoundSystem) PlaySound(name string, state *State) error {
	sound := state.decodeded_sounds[name]

	streamer := sound.streamer
	format := sound.format

	ctrl := &beep.Ctrl{Streamer: beep.Loop(1, streamer), Paused: false}

	sound_system.streamer = streamer
	sound_system.format = format
	sound_system.ctrl = ctrl

	beeper := beep.Seq(sound_system.ctrl, beep.Resample(4, format.SampleRate, 44100, ctrl))
	speaker.Play(beeper)

	return nil
}

func (sound_system *SoundSystem) Close() error {
	if sound_system.streamer != nil {
		if err := sound_system.streamer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func DecodeFile(path string) *DecodedSound {
	f, _ := os.Open(path)
	streamer, format, _ := mp3.Decode(f)
	return &DecodedSound{path, streamer, format}
}
