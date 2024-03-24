package main

import (
	"io"
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

func (sound_system *SoundSystem) PlaySound(file io.ReadCloser) error {
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}

	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}

	sound_system.streamer = streamer
	sound_system.format = format
	sound_system.ctrl = ctrl

	beeper := beep.Seq(sound_system.ctrl, beep.Resample(4, format.SampleRate, 44100, ctrl))

	if format.SampleRate != 0 {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		speaker.Play(beeper)
	}

	return nil
}
