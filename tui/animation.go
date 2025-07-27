package main

import (
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type Animation interface {
	tea.Model
	Tick()
	SetDimensions(width, height int)
}

// DefaultAnimation is to be embedded in other animations to define
// default behavior for SetDimensions and select tea.Model methods.
type DefaultAnimation struct {
	width  int
	height int

	parent Animation
}

func NewDefaultAnimation(parent Animation) *DefaultAnimation {
	return &DefaultAnimation{
		width:  80,
		height: 24,
		parent: parent,
	}
}

func (d *DefaultAnimation) SetDimensions(width, height int) {
	d.width = width
	d.height = height
	if d.width < 10 {
		d.width = 10 // Minimum width
	}
	if d.height < 10 {
		d.height = 10 // Minimum height
	}
}

func (d *DefaultAnimation) Init() tea.Cmd {
	return nil
}

func (d *DefaultAnimation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case timer.TickMsg:
		d.parent.Tick()
		return d.parent, nil
	}
	return d.parent, nil
}
