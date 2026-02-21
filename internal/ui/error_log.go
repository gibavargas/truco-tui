package ui

import tea "github.com/charmbracelet/bubbletea"

const maxUIErrorLog = 24

func (m *UIModel) setTransientError(err error) tea.Cmd {
	if err == nil {
		return nil
	}
	m.err = err
	line := tr("error_prefix") + err.Error()
	m.errorLog = append(m.errorLog, line)
	if len(m.errorLog) > maxUIErrorLog {
		m.errorLog = m.errorLog[len(m.errorLog)-maxUIErrorLog:]
	}
	m.visualState.errClearID++
	return clearErrorCmd(m.visualState.errClearID)
}
