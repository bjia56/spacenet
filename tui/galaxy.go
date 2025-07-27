package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type Galaxy struct {
	*DefaultAnimation

	numArms      int
	pointsPerArm int
	armSpread    float64
	spinSpeed    float64
	galaxyRadius float64
	angle        float64

	dotStyle  lipgloss.Style
	starStyle lipgloss.Style
	orbStyle  lipgloss.Style
	coreStyle lipgloss.Style
}

func NewGalaxy() *Galaxy {
	g := &Galaxy{
		numArms:      4,
		pointsPerArm: 60,
		armSpread:    0.5,
		spinSpeed:    0.12,
		galaxyRadius: 15.0,

		// Initialize lipgloss styles for spiral colors
		dotStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("33")),             // Blue
		starStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("36")),             // Cyan
		orbStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("226")),            // Yellow
		coreStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true), // Red
	}
	g.DefaultAnimation = NewDefaultAnimation(g)
	return g
}

func (g *Galaxy) Tick() {
	g.angle += g.spinSpeed
}

func (g *Galaxy) View() string {
	// Create empty screen (now [][]string for styled output)
	screen := make([][]string, g.height)
	for i := range screen {
		screen[i] = make([]string, g.width)
		for j := range screen[i] {
			screen[i][j] = " "
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
				// Vary brightness for spiral effect and add color using lipgloss
				var ch string
				if p < g.pointsPerArm/4 {
					ch = g.coreStyle.Render(".")
				} else if p < g.pointsPerArm/2 {
					ch = g.starStyle.Render("*")
				} else if p < 3*g.pointsPerArm/4 {
					ch = g.orbStyle.Render("*")
				} else {
					ch = g.dotStyle.Render(".")
				}
				screen[y][x] = ch
			}
		}
	}

	var output string
	for _, row := range screen {
		for _, ch := range row {
			output += ch
		}
		output += "\n"
	}
	return output
}
