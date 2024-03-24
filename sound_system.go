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
	defer streamer.Close()

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return err
	}

	speaker.Play(streamer)
	sound_system.streamer = streamer

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
