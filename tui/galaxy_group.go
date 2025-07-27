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
		satelliteCount:   60,    // More satellite galaxies
		orbitSpeed:       0.015, // Slower, more majestic motion
		interactStrength: 0.85,  // Stronger interactions
	}
	g.DefaultAnimation = NewDefaultAnimation(g)

	// Initialize major galaxies (enhanced Local Group representation)
	g.majorGalaxies = []MajorGalaxy{
		{ // Milky Way analog
			size:       1.2,
			orbitDist:  float64(g.height) * 0.28,
			orbitAngle: 0,
		},
		{ // Andromeda analog (M31)
			size:       1.5, // Larger than Milky Way
			orbitDist:  float64(g.height) * 0.28,
			orbitAngle: math.Pi,
		},
		{ // Triangulum analog (M33)
			size:       0.9,
			orbitDist:  float64(g.height) * 0.38,
			orbitAngle: math.Pi * 0.5,
		},
		{ // Large Magellanic Cloud analog
			size:       0.6,
			orbitDist:  float64(g.height) * 0.2,
			orbitAngle: math.Pi * 1.7,
		},
		{ // Small Magellanic Cloud analog
			size:       0.45,
			orbitDist:  float64(g.height) * 0.22,
			orbitAngle: math.Pi * 1.8,
		},
	}

	// Initialize styles with richer colors
	g.majorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("159")).Bold(true) // Bright cyan-white
	g.satelliteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("147"))        // Light purple
	g.streamStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("61"))            // Soft blue
	g.dustStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))             // Dark dust
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

		// Calculate galaxy center position with wider orbital paths
		aspectRatio := float64(g.width) / float64(g.height)
		galaxy.x = float64(cx) + math.Cos(galaxy.orbitAngle)*galaxy.orbitDist*aspectRatio*1.8
		galaxy.y = float64(cy) + math.Sin(galaxy.orbitAngle)*galaxy.orbitDist*0.9

		// Draw spiral arms with varying structure based on galaxy size
		arms := 4 + int(galaxy.size*2) // More arms for larger galaxies
		pointsPerArm := 35             // More points for denser arms

		// Calculate arm tightness based on galaxy size and screen dimensions
		screenScale := math.Min(float64(g.width), float64(g.height)) / 100.0
		armTightness := (0.3 + 0.15*math.Sin(galaxy.angle*0.5)) * screenScale

		for arm := 0; arm < arms; arm++ {
			// Base angle plus slight asymmetry
			armAngle := float64(arm)*2*math.Pi/float64(arms) + galaxy.angle +
				0.2*math.Sin(float64(arm)+galaxy.angle)

			for p := 0; p < pointsPerArm; p++ {
				progress := float64(p) / float64(pointsPerArm)

				// Radius with enhanced arm length and perturbations
				baseRadius := math.Min(float64(g.width), float64(g.height)) / 4.2
				r := progress * galaxy.size * baseRadius

				// Add more complex spiral arm structure
				armWave := 0.2 * math.Sin(progress*4+galaxy.angle) * (1.0 - progress*0.5)
				spiralTwist := math.Pow(progress, 0.7) // Non-linear spiral winding
				theta := armAngle + r*armTightness*spiralTwist + armWave

				// Calculate position with enhanced elliptical distortion
				aspectScale := float64(g.width) / float64(g.height)
				x := int(galaxy.x + r*math.Cos(theta)*math.Min(aspectScale, 1.2))
				y := int(galaxy.y + r*math.Sin(theta)*0.6)

				if x >= 0 && x < g.width && y >= 0 && y < g.height {
					// Vary the appearance based on position and galaxy size
					if p < pointsPerArm/4 {
						screen[y][x] = g.majorStyle.Render("@") // Bright core
					} else if p < pointsPerArm/2 {
						screen[y][x] = g.majorStyle.Render("*") // Inner arms
					} else {
						screen[y][x] = g.majorStyle.Render("·") // Outer regions
					}
				}
			}
		}

		// Draw tidal streams between major galaxies
		if i > 0 {
			prev := g.majorGalaxies[i-1]
			steps := 15

			// Calculate interaction strength based on galaxy sizes and distances
			dx, dy := galaxy.x-prev.x, galaxy.y-prev.y
			dist := math.Sqrt(dx*dx + dy*dy)
			interaction := g.interactStrength * (galaxy.size + prev.size) / (dist + 1)

			for s := 0; s < steps; s++ {
				progress := float64(s) / float64(steps)

				// Complex wave pattern based on interaction strength
				wave1 := math.Sin(progress*math.Pi*2+g.offset) * float64(g.height/8)
				wave2 := math.Cos(progress*math.Pi*3+g.offset*0.7) * float64(g.height/12)
				finalWave := (wave1 + wave2) * interaction

				// Calculate stream position with gravitational curvature
				x := int(prev.x + (galaxy.x-prev.x)*progress)
				y := int(prev.y + (galaxy.y-prev.y)*progress + finalWave)

				// Draw wider streams near galaxies
				streamWidth := int(2 * (1 - math.Abs(progress-0.5)))

				for w := -streamWidth; w <= streamWidth; w++ {
					drawX := x + w
					drawY := y + w/2

					if drawX >= 0 && drawX < g.width && drawY >= 0 && drawY < g.height {
						if screen[drawY][drawX] == " " {
							if math.Abs(float64(w)) < float64(streamWidth)/2 {
								screen[drawY][drawX] = g.streamStyle.Render("∙")
							} else {
								screen[drawY][drawX] = g.dustStyle.Render("·")
							}
						}
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

		// Calculate unique orbital parameters for each satellite
		baseFreq := 1.0 + float64(i%3)*0.2
		satAngle := float64(i)*2*math.Pi/float64(g.satelliteCount) +
			g.offset*baseFreq +
			math.Sin(g.offset*0.5+float64(i))*0.2 // Orbital perturbations

		// Enhanced satellite distribution with screen-aware scaling
		screenScale := math.Min(float64(g.width), float64(g.height)) / 6
		baseDist := screenScale * major.size

		// More varied orbital distances
		satDist := baseDist * (0.8 +
			math.Sin(float64(i)*1.7)*0.3 + // Static variation
			math.Sin(g.offset*0.7+float64(i))*0.2 + // Dynamic variation
			math.Cos(float64(i)*0.5)*0.15) // Additional orbital diversity

		// Calculate elliptical orbit with aspect ratio consideration
		aspectRatio := float64(g.width) / float64(g.height)
		x := int(major.x + math.Cos(satAngle)*satDist*math.Min(aspectRatio, 1.4))
		y := int(major.y + math.Sin(satAngle)*satDist*0.7)

		if x >= 0 && x < g.width && y >= 0 && y < g.height {
			// Vary satellite appearance based on position
			if i%5 == 0 {
				screen[y][x] = g.satelliteStyle.Render("○") // Larger satellites
			} else {
				screen[y][x] = g.satelliteStyle.Render("·") // Small satellites
			}
		}

		// Enhanced dust trails with varying density
		if i%4 == 0 { // More frequent trails
			trailLength := 4 + int(math.Sin(float64(i))*2)
			for d := 1; d < trailLength; d++ {
				// Calculate trail position with curved path
				trailProgress := float64(d) / float64(trailLength)
				trailAngle := satAngle - trailProgress*0.3
				trailDist := satDist * (1.0 - trailProgress*0.2)

				trailX := int(major.x + math.Cos(trailAngle)*trailDist*1.2)
				trailY := int(major.y + math.Sin(trailAngle)*trailDist*0.6)

				if trailX >= 0 && trailX < g.width && trailY >= 0 && trailY < g.height {
					if screen[trailY][trailX] == " " {
						if d < 2 {
							screen[trailY][trailX] = g.dustStyle.Render("∙")
						} else {
							screen[trailY][trailX] = g.dustStyle.Render("·")
						}
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
