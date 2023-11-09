package edit

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

type Session struct {
	Input  string
	Output string
}

func (s *Session) Edit() error {
	// prepare input
	temp, err := os.CreateTemp("", "forester-snippet-XXXXXXX.ks")
	if err != nil {
		return fmt.Errorf("cannot create temporary file: %w", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Printf("error closing temp file: %s", err.Error())
		}
	}(temp.Name())
	err = os.WriteFile(temp.Name(), []byte(s.Input), 0x600)
	if err != nil {
		return fmt.Errorf("cannot create snippet: %w", err)
	}

	// detect editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = default_editor
	}
	editorPath, err := exec.LookPath(editor)
	if err != nil {
		return fmt.Errorf("editor unavailable: %w", err)
	}

	// open editor
	cmd := exec.Command(editorPath, temp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing editor: %w", err)
	}

	// on windows platforms wait until file is saved
	if wait_for_editor {
		fmt.Print("Press 'Enter' after the changes are saved...")
		_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	// read the result
	buf, err := os.ReadFile(temp.Name())
	if err != nil {
		return fmt.Errorf("cannot read temp file: %w", err)
	}
	s.Output = string(buf)

	return nil
}
