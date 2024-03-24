package main

import (
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type SoundSystem struct {
	streamer beep.StreamSeekCloser
	format   beep.Format
	ctrl     *beep.Ctrl
}

func NewSoundSystem() *SoundSystem {
	return &SoundSystem{}
}

func (sound_system *SoundSystem) PlaySound(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}

	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}

	sound_system.streamer = streamer
	sound_system.format = format
	sound_system.ctrl = ctrl

	beeper := beep.Seq(sound_system.ctrl, beep.Resample(4, format.SampleRate, 44100, ctrl))

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beeper)

	// Calculate the duration of the sound
	duration := float64(streamer.Len()) / float64(format.SampleRate.N(time.Second))

	// Wait for the sound to finish
	time.Sleep(time.Duration(duration) * time.Second)

	// Close the sound
	if err := sound_system.Close(); err != nil {
		return err
	}
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
