package main

import (
	"math"
	"time"
	"unsafe"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type Galaxy struct {
	numArms      int
	pointsPerArm int
	armSpread    float64
	spinSpeed    float64
	galaxyRadius float64
	angle        float64

	width  int
	height int
}

func (g *Galaxy) Initialize() {
	g.numArms = 4
	g.pointsPerArm = 60
	g.armSpread = 0.5
	g.spinSpeed = 0.12
	g.galaxyRadius = 9.0
	g.width = 80
	g.height = 24
}

func (g *Galaxy) SetDimensions(width, height int) {
	g.width = width
	g.height = height
	if g.width < 10 {
		g.width = 10 // Minimum width
	}
	if g.height < 10 {
		g.height = 10 // Minimum height
	}
}

func (g *Galaxy) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timer.TickMsg{ID: int(uintptr(unsafe.Pointer(g)))}
	})
}

func (g *Galaxy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case timer.TickMsg:
		g.Tick()
		return g, g.Init()
	}
	return g, nil
}

func (g *Galaxy) Tick() {
	g.angle += g.spinSpeed
}

func (g *Galaxy) View() string {
	// Create empty screen
	screen := make([][]rune, g.height)
	for i := range screen {
		screen[i] = make([]rune, g.width)
		for j := range screen[i] {
			screen[i][j] = ' '
		}
	}

	cx, cy := g.width/2, g.height/2

	for arm := 0; arm < g.numArms; arm++ {
		armAngle := float64(arm) * 2 * math.Pi / float64(g.numArms)
		for p := 0; p < g.pointsPerArm; p++ {
			r := float64(p) / float64(g.pointsPerArm) * g.galaxyRadius
			theta := armAngle + g.angle + r*g.armSpread

			// Add a slight spiral
			x := int(float64(cx) + r*math.Cos(theta))
			y := int(float64(cy) + r*math.Sin(theta)*0.5) // squash for ellipse

			if x >= 0 && x < g.width && y >= 0 && y < g.height {
				// Vary brightness for spiral effect
				var ch rune
				if p < g.pointsPerArm/4 {
					ch = '.'
				} else if p < g.pointsPerArm/2 {
					ch = '*'
				} else if p < 3*g.pointsPerArm/4 {
					ch = 'o'
				} else {
					ch = '@'
				}
				screen[y][x] = ch
			}
		}
	}

	var output string
	for _, row := range screen {
		output += string(row) + "\n"
	}
	return output
}
