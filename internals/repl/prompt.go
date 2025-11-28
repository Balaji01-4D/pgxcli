package repl

func RunRepl(db string) (string, error) {
	m := NewModel(db)

	input, err := m.GetLine()
	if err != nil {
		return "", err
	}

	// p := tea.NewProgram(m)

	// if _, err := p.Run(); err != nil {
	// 	return err
	// }
	// return nil
	return input, nil
}
