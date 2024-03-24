package main

import (
	"encoding/csv"
	"log"
	"strings"

	_ "embed"
)

//go:embed res/dialogues.csv
var dialoguesCsv string

var soundSystem *SoundSystem

type Dialog struct {
	name  string
	value string
}

func NewDialog(name, value string) *Dialog {
	return &Dialog{
		name:  name,
		value: value,
	}
}

func getDialogues() []*Dialog {
	// Parse the CSV
	records, err := csv.NewReader(strings.NewReader(dialoguesCsv)).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Create the dialogues
	dialogues := make([]*Dialog, 0)
	for i := 1; i < len(records); i++ {
		record := records[i]
		dialogues = append(dialogues, NewDialog(record[0], record[1]))
	}

	return dialogues
}

func displayDialogue(dialogues []*Dialog, name string, state *State) {
	dialogue := getDialogue(dialogues, name)

	if text, err := NewText(state, dialogue, 0, 0, 1, 1); err != nil {
		panic(err)
	} else {
		state.text = text
	}

	if soundSystem == nil {
		soundSystem = NewSoundSystem()
		soundSystem.InitSpeaker(state.decodeded_sounds[name].format)
	}

	if err := soundSystem.PlaySound(name, state); err != nil {
		panic(err)
	}
}

func getDialogue(dialogues []*Dialog, name string) string {
	for _, dialogue := range dialogues {
		if dialogue.name == name {
			return dialogue.value
		}
	}

	return ""
}
