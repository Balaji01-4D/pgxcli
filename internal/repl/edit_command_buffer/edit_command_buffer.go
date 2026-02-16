package editcommandbuffer

import (
	"errors"
	"io"
	"os"
	"os/exec"
)

var ErrEditorNotFound = errors.New("Editor not found, make sure envirnoment variable is applied for $VISUAL or $EDITOR")

type EditCommandBuffer struct {
	currentInput string
	editor       string
}

func New(currentInput string) (*EditCommandBuffer, error) {
	editor, err := getEditorFromEnv()
	if err != nil {
		return nil, err
	}

	return &EditCommandBuffer{
		currentInput: currentInput,
		editor:       editor,
	}, nil
}

func (e *EditCommandBuffer) Run() (string, error) {
	tempFile, err := os.CreateTemp("", "pgxcli-*.sql")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	if _, wErr := tempFile.WriteString(e.currentInput); wErr != nil {
		return "", err
	}
	tempFile.Close()

	if editorErr := runEditor(e.editor, tempFile.Name()); editorErr != nil {
		return "", editorErr
	}

	rf, err := os.Open(tempFile.Name())
	if err != nil {
		return "", err
	}
	defer rf.Close()

	data, err := io.ReadAll(rf)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func runEditor(editor, file string) error {
	cmd := exec.Command(editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func getEditorFromEnv() (string, error) {
	if editor, exist := os.LookupEnv("VISUAL"); exist {
		return editor, nil
	} else if editor, exist := os.LookupEnv("EDITOR"); exist {
		return editor, nil
	} else {
		return "", ErrEditorNotFound
	}
}
