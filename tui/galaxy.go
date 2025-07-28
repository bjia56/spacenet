package main

import (
	"fmt"
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
	isClockwise  bool        // Direction of spiral rotation

	// Visual styles for different elements
	coreStyle lipgloss.Style // Galaxy core
	starStyle lipgloss.Style // Bright stars
	armStyle  lipgloss.Style // Main arm stars
	dustStyle lipgloss.Style // Dust and faint stars

	// Background particle field parameters
	particlePositions [][2]float64 // Normalized positions (0-1)
	particlePhases    []float64    // Phase offsets for particle movement
	particleOrbits    []float64    // Orbital parameters for particles
	particleTypes     []byte       // Type of each particle (for visual variation)
}

func NewGalaxy() *Galaxy {
	g := &Galaxy{
		size:      1.0,  // Will be randomized in ResetParameters
		spinSpeed: 0.12, // Will be randomized in ResetParameters
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
	g.isClockwise = randParams[4]%2 == 0 // Random rotation direction

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

	// Initialize color styles based on galaxy characteristics
	colorBytes := g.RandBytes(4) // Get random bytes for color variations

	// Core color - range from red to orange to yellow based on galaxy size
	// Larger galaxies are hotter (more yellow), smaller are redder
	var coreBase int
	if g.size > 1.1 {
		coreBase = 220 + int(colorBytes[0]%3) // 220-222: warm yellows
	} else if g.size > 0.9 {
		coreBase = 214 + int(colorBytes[0]%3) // 214-216: oranges
	} else {
		coreBase = 196 + int(colorBytes[0]%3) // 196-198: reds
	}
	g.coreStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", coreBase))).
		Bold(true)

	// Star colors - influenced by core temperature and arm count
	var starBase int
	if g.numArms > 5 {
		starBase = 229 + int(colorBytes[1]%3) // 229-231: bright whites
	} else if g.numArms > 4 {
		starBase = 226 + int(colorBytes[1]%3) // 226-228: warm yellows
	} else {
		starBase = 223 + int(colorBytes[1]%3) // 223-225: soft yellows
	}
	g.starStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", starBase)))

	// Arm colors - vary based on size and arm tightness
	var armBase int
	if g.armTightness > 0.4 {
		armBase = 159 + int(colorBytes[2]%3) // 159-161: bright blue-whites
	} else if g.armTightness > 0.3 {
		armBase = 153 + int(colorBytes[2]%3) // 153-155: cyan-whites
	} else {
		armBase = 147 + int(colorBytes[2]%3) // 147-149: soft blues
	}
	g.armStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", armBase)))

	// Dust colors - deeper and darker variations
	dustOptions := []string{"236", "237", "238", "239"}
	dustIndex := int(colorBytes[3]) % len(dustOptions)
	g.dustStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(dustOptions[dustIndex]))

	// Initialize background particle field with fixed number of particles
	const baseParticles = 150 // Base number of particles for a single galaxy
	g.particlePositions = make([][2]float64, baseParticles)
	g.particlePhases = make([]float64, baseParticles)
	g.particleOrbits = make([]float64, baseParticles)
	g.particleTypes = make([]byte, baseParticles)

	// Get random bytes for particle initialization
	particleBytes := g.RandBytes(baseParticles * 4) // 4 bytes per particle (x, y, orbit, type)
	for i := range baseParticles {
		// Convert to polar coordinates for better orbital distribution
		angle := float64(particleBytes[i*4]) / 255.0 * math.Pi * 2
		radius := 0.1 + float64(particleBytes[i*4+1])/255.0*0.9 // Radial distribution

		// Store in normalized coordinates (0-1 range)
		g.particlePositions[i][0] = 0.5 + math.Cos(angle)*radius*0.5
		g.particlePositions[i][1] = 0.5 + math.Sin(angle)*radius*0.5

		// Phase and orbit parameters
		g.particlePhases[i] = float64(particleBytes[i*4+2]) / 255.0 * math.Pi * 2
		g.particleOrbits[i] = 0.5 + float64(particleBytes[i*4+3])/255.0 // Orbit speed multiplier
		g.particleTypes[i] = particleBytes[i*4+3]
	}
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

	// Draw background particle field
	screenArea := g.width * g.height
	particleDensity := float64(screenArea) / 10000.0 // Adjust density for single galaxy scale

	for i, pos := range g.particlePositions {
		// Skip some particles based on screen size to maintain consistent density
		if float64(i) > float64(len(g.particlePositions))*particleDensity {
			break
		}

		// Calculate radial distance from center
		dx := pos[0] - 0.5
		dy := pos[1] - 0.5
		distFromCenter := math.Sqrt(dx*dx+dy*dy) * 2.0

		// Calculate orbital motion
		phase := g.particlePhases[i]
		orbitSpeed := g.particleOrbits[i] * g.spinSpeed * 0.5

		// Complex orbital motion
		baseAngle := math.Atan2(dy, dx)
		orbitDir := 1.0
		if !g.isClockwise {
			orbitDir = -1.0
		}
		newAngle := baseAngle + orbitDir*orbitSpeed*(1.0-distFromCenter*0.3) // Faster orbits near center
		newRadius := math.Sqrt(dx*dx+dy*dy) * (1.0 + math.Sin(g.angle+phase)*0.1)

		// Calculate screen position
		screenX := int(float64(cx) + newRadius*math.Cos(newAngle)*float64(g.width))
		screenY := int(float64(cy) + newRadius*math.Sin(newAngle)*float64(g.height))

		// Only draw if in bounds and not overlapping
		if screenX >= 0 && screenX < g.width && screenY >= 0 && screenY < g.height && screen[screenY][screenX] == " " {
			particleType := g.particleTypes[i]

			if distFromCenter < 0.3 && particleType%5 == 0 {
				// Particles in the dense central region
				screen[screenY][screenX] = g.coreStyle.Render("∙")
			} else if particleType%7 == 0 {
				// Occasional brighter particles
				screen[screenY][screenX] = g.starStyle.Render("*")
			} else {
				// Background dust
				screen[screenY][screenX] = g.dustStyle.Render("∙")
			}
		}
	}

	// Calculate screen-aware scaling for spiral arms
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

			// Apply spiral direction
			twistDir := 1.0
			if !g.isClockwise {
				twistDir = -1.0
			}
			theta := armAngle + twistDir*r*g.armTightness*spiralTwist + armWave

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
