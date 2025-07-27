package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type GalaxyGroup struct {
	*DefaultAnimation

	majorGalaxies    []MajorGalaxy // Major galaxies in the group
	satelliteCount   int           // Number of satellite galaxies
	orbitSpeed       float64       // Base orbital speed
	interactStrength float64       // Strength of gravitational interactions
	offset           float64       // Animation offset

	// Visual styles
	majorStyle     lipgloss.Style // Major galaxies
	satelliteStyle lipgloss.Style // Satellite galaxies
	streamStyle    lipgloss.Style // Tidal streams
	dustStyle      lipgloss.Style // Interstellar dust
}

type MajorGalaxy struct {
	x, y       float64 // Position
	size       float64 // Relative size
	angle      float64 // Current rotation angle
	orbitDist  float64 // Distance from group center
	orbitAngle float64 // Angle in orbit
}

func NewGalaxyGroup() *GalaxyGroup {
	g := &GalaxyGroup{
		satelliteCount:   40,
		orbitSpeed:       0.02,
		interactStrength: 0.7,
	}
	g.DefaultAnimation = NewDefaultAnimation(g)

	// Initialize major galaxies (similar to Local Group)
	g.majorGalaxies = []MajorGalaxy{
		{ // Milky Way analog
			size:       1.0,
			orbitDist:  float64(g.height) * 0.2,
			orbitAngle: 0,
		},
		{ // Andromeda analog
			size:       1.2,
			orbitDist:  float64(g.height) * 0.2,
			orbitAngle: math.Pi,
		},
		{ // Triangulum analog
			size:       0.6,
			orbitDist:  float64(g.height) * 0.3,
			orbitAngle: math.Pi * 0.5,
		},
	}

	// Initialize styles
	g.majorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))      // Bright white
	g.satelliteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246")) // Gray
	g.streamStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))    // Dark gray
	g.dustStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))      // Medium gray
	return g
}

func (g *GalaxyGroup) Tick() {
	g.offset += g.orbitSpeed

	// Update major galaxy positions
	for i := range g.majorGalaxies {
		galaxy := &g.majorGalaxies[i]
		galaxy.orbitAngle += g.orbitSpeed * (1.0 / (1.0 + float64(i)*0.5))
		galaxy.angle += g.orbitSpeed * 2
	}
}

func (g *GalaxyGroup) View() string {
	screen := make([][]string, g.height)
	for i := range screen {
		screen[i] = make([]string, g.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	cx, cy := g.width/2, g.height/2

	// Update and draw major galaxies
	for i := range g.majorGalaxies {
		galaxy := &g.majorGalaxies[i]

		// Calculate galaxy center position
		galaxy.x = float64(cx) + math.Cos(galaxy.orbitAngle)*galaxy.orbitDist*2.0
		galaxy.y = float64(cy) + math.Sin(galaxy.orbitAngle)*galaxy.orbitDist

		// Draw spiral arms for each major galaxy
		arms := 4
		pointsPerArm := 20
		for arm := 0; arm < arms; arm++ {
			armAngle := float64(arm)*2*math.Pi/float64(arms) + galaxy.angle
			for p := 0; p < pointsPerArm; p++ {
				r := float64(p) / float64(pointsPerArm) * galaxy.size * float64(g.height/6)
				theta := armAngle + r*0.5

				x := int(galaxy.x + r*math.Cos(theta))
				y := int(galaxy.y + r*math.Sin(theta)*0.5)

				if x >= 0 && x < g.width && y >= 0 && y < g.height {
					if p < pointsPerArm/3 {
						screen[y][x] = g.majorStyle.Render("*")
					} else {
						screen[y][x] = g.majorStyle.Render("·")
					}
				}
			}
		}

		// Draw tidal streams between major galaxies
		if i > 0 {
			prev := g.majorGalaxies[i-1]
			steps := 10
			for s := 0; s < steps; s++ {
				progress := float64(s) / float64(steps)
				// Add wave effect to the streams
				wave := math.Sin(progress*math.Pi*2+g.offset) * float64(g.height/10)

				x := int(prev.x + (galaxy.x-prev.x)*progress)
				y := int(prev.y + (galaxy.y-prev.y)*progress + wave)

				if x >= 0 && x < g.width && y >= 0 && y < g.height {
					if screen[y][x] == " " {
						screen[y][x] = g.streamStyle.Render("∙")
					}
				}
			}
		}
	}

	// Draw satellite galaxies
	for i := 0; i < g.satelliteCount; i++ {
		// Choose which major galaxy to orbit
		majorIndex := i % len(g.majorGalaxies)
		major := g.majorGalaxies[majorIndex]

		// Calculate satellite position
		satAngle := float64(i)*2*math.Pi/float64(g.satelliteCount) + g.offset*(1.0+float64(i)*0.1)
		satDist := float64(g.height/8) * (0.5 + math.Sin(float64(i))*0.3)

		x := int(major.x + math.Cos(satAngle)*satDist)
		y := int(major.y + math.Sin(satAngle)*satDist*0.5)

		if x >= 0 && x < g.width && y >= 0 && y < g.height {
			screen[y][x] = g.satelliteStyle.Render(".")
		}

		// Add dust trails for some satellites
		if i%3 == 0 {
			for d := 1; d < 3; d++ {
				trailX := int(major.x + math.Cos(satAngle-float64(d)*0.2)*satDist)
				trailY := int(major.y + math.Sin(satAngle-float64(d)*0.2)*satDist*0.5)

				if trailX >= 0 && trailX < g.width && trailY >= 0 && trailY < g.height {
					if screen[trailY][trailX] == " " {
						screen[trailY][trailX] = g.dustStyle.Render("·")
					}
				}
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
