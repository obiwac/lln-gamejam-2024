package main

import (
	_ "embed"
)

type WorlObama struct {
	World
}

func NewWorldObama(state *State) (*WorlObama, error) {
	obama := &WorlObama{}
	obama.World = World{state: state}

	return obama, nil
}

func (world *WorlObama) Render() {
}

func (world *WorlObama) Release() {
}
