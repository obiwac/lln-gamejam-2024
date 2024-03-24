package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"log"
	"strings"

	_ "embed"
)

//go:embed res/dialogues.csv
var dialoguesCsv string

//go:embed res/sound/intro1.mp3
var intro1 []byte

//go:embed res/sound/intro2.mp3
var intro2 []byte

//go:embed res/sound/outro1.mp3
var outro1 []byte

//go:embed res/sound/outro4.mp3
var outro4 []byte

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

	soundFile := io.ReadCloser(nil)
	switch name {
	case "intro1":
		soundFile = bytes.NewReader(intro1)
	case "intro2":
		soundFile = bytes.NewReader(intro2)
	case "outro1":
		soundFile = bytes.NewReader(outro1)
	case "outro4":
		soundFile = bytes.NewReader(outro4)
	}

	NewSoundSystem().PlaySound(soundFile)
}

func getDialogue(dialogues []*Dialog, name string) string {
	for _, dialogue := range dialogues {
		if dialogue.name == name {
			return dialogue.value
		}
	}

	return ""
}
