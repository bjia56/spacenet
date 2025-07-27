package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type Galaxy struct {
	*DefaultAnimation

	size      float64 // Overall size multiplier
	angle     float64 // Current rotation angle
	spinSpeed float64 // Rotation speed

	// Stored random parameters for deterministic rendering
	armSeeds     [][]float64 // Random seeds for spiral arm variations
	starSeeds    []float64   // Random seeds for star brightness
	dustPatterns []float64   // Random patterns for dust distribution
	pointsPerArm int         // Number of points per arm
	numArms      int         // Number of arms
	armTightness float64     // Base arm tightness

	// Visual styles for different elements
	coreStyle lipgloss.Style // Galaxy core
	starStyle lipgloss.Style // Bright stars
	armStyle  lipgloss.Style // Main arm stars
	dustStyle lipgloss.Style // Dust and faint stars
}

func NewGalaxy() *Galaxy {
	g := &Galaxy{
		size:      1.0,  // Will be randomized in ResetParameters
		spinSpeed: 0.12, // Will be randomized in ResetParameters

		// Initialize styles with rich color palette
		coreStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true), // Red core
		starStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("226")),            // Yellow bright stars
		armStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("159")),            // Cyan-white arm stars
		dustStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("238")),            // Dark dust
	}
	g.DefaultAnimation = NewDefaultAnimation(g)
	return g
}

func (g *Galaxy) Tick() {
	g.angle += g.spinSpeed
}

func (g *Galaxy) ResetParameters() {
	// Get random bytes for initial parameters
	randParams := g.RandBytes(8)

	// Randomize core parameters
	g.numArms = 4 + int(randParams[0]%3) // 4-6 arms
	g.pointsPerArm = 45                  // More points for denser arms
	g.spinSpeed = 0.012 + float64(randParams[1]%10)/1000.0
	g.size = 0.8 + float64(randParams[2])/255.0*0.4 // 0.8-1.2 size multiplier
	g.armTightness = 0.3 + float64(randParams[3])/255.0*0.15

	// Initialize arm seeds for each arm with two random values per arm
	g.armSeeds = make([][]float64, g.numArms)
	for i := range g.armSeeds {
		seeds := g.RandBytes(2) // 2 random values per arm
		g.armSeeds[i] = make([]float64, 2)
		for j := range seeds {
			g.armSeeds[i][j] = float64(seeds[j]) / 255.0
		}
	}

	// Initialize star brightness variations
	g.starSeeds = make([]float64, g.pointsPerArm)
	starBytes := g.RandBytes(g.pointsPerArm)
	for i := range g.starSeeds {
		g.starSeeds[i] = float64(starBytes[i]) / 255.0
	}

	// Initialize dust pattern variations
	g.dustPatterns = make([]float64, g.pointsPerArm)
	dustBytes := g.RandBytes(g.pointsPerArm)
	for i := range g.dustPatterns {
		g.dustPatterns[i] = float64(dustBytes[i]) / 255.0
	}

	// Reset animation state
	g.angle = float64(randParams[4]) / 255.0 * math.Pi * 2 // Random initial rotation
}

func (g *Galaxy) View() string {
	screen := make([][]string, g.height)
	for i := range screen {
		screen[i] = make([]string, g.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	cx, cy := g.width/2, g.height/2

	// Calculate screen-aware scaling
	baseRadius := math.Min(float64(g.width), float64(g.height)) / 2
	aspectRatio := float64(g.width) / float64(g.height)

	for arm := 0; arm < g.numArms; arm++ {
		// Use stored random values for arm variation
		armSeedBase := g.armSeeds[arm][0]
		armSeedTwist := g.armSeeds[arm][1]

		// Base angle with deterministic asymmetry
		armAngle := float64(arm)*2*math.Pi/float64(g.numArms) + g.angle +
			0.3*armSeedBase*math.Sin(float64(arm)+g.angle)

		for p := 0; p < g.pointsPerArm; p++ {
			progress := float64(p) / float64(g.pointsPerArm)

			// Radius with enhanced arm length
			r := progress * g.size * baseRadius

			// Add complex spiral arm structure with stored randomness
			armWave := 0.2 * (0.8 + 0.4*armSeedBase) *
				math.Sin(progress*4+g.angle) * (1.0 - progress*0.5)
			spiralTwist := math.Pow(progress, 0.6+0.2*armSeedTwist)
			theta := armAngle + r*g.armTightness*spiralTwist + armWave

			// Calculate position with enhanced elliptical distortion
			x := int(float64(cx) + r*math.Cos(theta)*math.Min(aspectRatio, 1.2))
			y := int(float64(cy) + r*math.Sin(theta)*0.6)

			if x >= 0 && x < g.width && y >= 0 && y < g.height {
				var ch string
				starBright := g.starSeeds[p]
				dustDensity := g.dustPatterns[p]

				if p < g.pointsPerArm/6 {
					// Dense core region
					if starBright > 0.7 {
						ch = g.coreStyle.Render("@")
					} else {
						ch = g.starStyle.Render("*")
					}
				} else if p < g.pointsPerArm/2 {
					// Spiral arms
					if starBright > 0.8 {
						ch = g.starStyle.Render("*") // Bright stars
					} else if starBright > 0.5 {
						ch = g.armStyle.Render("*") // Normal stars
					} else {
						ch = g.armStyle.Render("·") // Dim stars
					}
				} else {
					// Outer regions
					if dustDensity > 0.95 {
						ch = g.starStyle.Render("*") // Rare bright stars
					} else if dustDensity > 0.8 {
						ch = g.armStyle.Render("·") // Dust lanes
					} else if dustDensity > 0.6 {
						ch = g.dustStyle.Render("·") // Faint dust
					} else {
						ch = g.dustStyle.Render("∙") // Very faint dust
					}
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
