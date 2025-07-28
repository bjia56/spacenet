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

	// Stored random parameters for deterministic rendering
	armSeeds        [][]float64 // Random seeds for spiral arm variations per galaxy
	satelliteOrbits []float64   // Random orbital parameters for satellites
	satellitePhases []float64   // Random phase shifts for satellites
	trailLengths    []int       // Random dust trail lengths

	// Background particle field parameters
	particlePositions [][2]float64 // Pre-calculated particle positions
	particlePhases    []float64    // Phase offsets for particle movement
	particleOrbits    []float64    // Orbital parameters for particles
	particleTypes     []byte       // Type of each particle (for visual variation)

	// Visual styles
	majorStyle     lipgloss.Style // Major galaxies
	satelliteStyle lipgloss.Style // Satellite galaxies
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

	// Initialize major galaxies with normalized distances (relative to 1.0)
	g.majorGalaxies = []MajorGalaxy{
		{ // Milky Way analog
			size:       1.2,
			orbitDist:  0.18, // Normalized distance
			orbitAngle: 0,
		},
		{ // Andromeda analog (M31)
			size:       1.5,
			orbitDist:  0.18,
			orbitAngle: math.Pi,
		},
		{ // Triangulum analog (M33)
			size:       0.9,
			orbitDist:  0.28,
			orbitAngle: math.Pi * 0.5,
		},
		{ // Large Magellanic Cloud analog
			size:       0.6,
			orbitDist:  0.10,
			orbitAngle: math.Pi * 1.7,
		},
		{ // Small Magellanic Cloud analog
			size:       0.45,
			orbitDist:  0.12,
			orbitAngle: math.Pi * 1.8,
		},
	}

	// Initialize styles with richer colors
	g.majorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("159")).Bold(true) // Bright cyan-white
	g.satelliteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("147"))        // Light purple
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

func (g *GalaxyGroup) ResetParameters() {
	// Get random bytes for initial parameters
	randParams := g.RandBytes(8)

	// Set core parameters with random variations
	g.satelliteCount = 50 + int(randParams[0]%30) // 50-79 satellites
	g.orbitSpeed = 0.012 + float64(randParams[1]%10)/1000.0
	g.interactStrength = 0.7 + float64(randParams[2]%40)/100.0

	// Determine number of major galaxies (4-7)
	numGalaxies := 4 + int(randParams[3]%4)

	// Base templates for galaxy types
	baseTemplates := []struct {
		baseSize float64
		baseDist float64
	}{
		{1.5, 0.18},  // Large spiral
		{1.2, 0.20},  // Medium spiral
		{0.9, 0.22},  // Small spiral
		{0.6, 0.15},  // Dwarf galaxy
		{0.45, 0.12}, // Small dwarf
	}

	// Get random bytes for galaxy placement
	galaxyParams := g.RandBytes(numGalaxies * 3) // Size variation, distance variation, angle

	// Initialize galaxies with random variations
	g.majorGalaxies = make([]MajorGalaxy, numGalaxies)
	galaxyVariations := g.RandBytes(numGalaxies * 2)

	for i := 0; i < numGalaxies; i++ {
		// Pick a random template for this galaxy
		templateIdx := int(galaxyParams[i*3] % uint8(len(baseTemplates)))
		template := baseTemplates[templateIdx]

		// Calculate variations
		sizeVar := float64(galaxyVariations[i*2]) / 255.0 * 0.2
		distVar := float64(galaxyVariations[i*2+1]) / 255.0 * 0.1

		// Calculate random angle for initial placement
		angle := float64(galaxyParams[i*3+2]) / 255.0 * math.Pi * 2

		g.majorGalaxies[i] = MajorGalaxy{
			size:       template.baseSize * (0.9 + sizeVar),              // ±10% size variation
			orbitDist:  template.baseDist * (1.0 + distVar),              // ±10% distance variation
			orbitAngle: angle,                                            // Random initial angle
			angle:      float64(randParams[3+i%5]) / 255.0 * math.Pi * 2, // Random initial rotation
		}
	}

	// Initialize arm seeds for each galaxy
	g.armSeeds = make([][]float64, len(g.majorGalaxies))
	for i := range g.majorGalaxies {
		armCount := 4 + int(g.majorGalaxies[i].size*2)
		seeds := g.RandBytes(armCount * 2) // 2 random values per arm
		g.armSeeds[i] = make([]float64, armCount*2)
		for j := range seeds {
			g.armSeeds[i][j] = float64(seeds[j]) / 255.0
		}
	}

	// Initialize satellite parameters
	g.satelliteOrbits = make([]float64, g.satelliteCount)
	g.satellitePhases = make([]float64, g.satelliteCount)
	g.trailLengths = make([]int, g.satelliteCount)

	satBytes := g.RandBytes(g.satelliteCount * 3)
	for i := 0; i < g.satelliteCount; i++ {
		g.satelliteOrbits[i] = 0.7 + float64(satBytes[i])/255.0*0.6 // 0.7-1.3 orbit scale
		g.satellitePhases[i] = float64(satBytes[i+g.satelliteCount]) / 255.0 * math.Pi * 2
		g.trailLengths[i] = 3 + int(satBytes[i+g.satelliteCount*2]%4) // 3-6 length
	}

	// Initialize background particle field with fixed number of particles
	const baseParticles = 300 // Base number of particles that will scale with screen size
	g.particlePositions = make([][2]float64, baseParticles)
	g.particlePhases = make([]float64, baseParticles)
	g.particleOrbits = make([]float64, baseParticles)
	g.particleTypes = make([]byte, baseParticles)

	// Get random bytes for particle initialization
	particleBytes := g.RandBytes(baseParticles * 4) // 4 bytes per particle (x, y, orbit, type)
	for i := 0; i < baseParticles; i++ {
		// Convert to polar coordinates for better orbital distribution
		angle := float64(particleBytes[i*4]) / 255.0 * math.Pi * 2
		radius := 0.1 + float64(particleBytes[i*4+1])/255.0*0.9 // Radial distribution

		// Store in normalized cartesian coordinates
		g.particlePositions[i][0] = 0.5 + math.Cos(angle)*radius*0.5
		g.particlePositions[i][1] = 0.5 + math.Sin(angle)*radius*0.5

		// Phase and orbit parameters
		g.particlePhases[i] = float64(particleBytes[i*4+2]) / 255.0 * math.Pi * 2
		g.particleOrbits[i] = 0.5 + float64(particleBytes[i*4+3])/255.0 // Orbit speed multiplier
		g.particleTypes[i] = particleBytes[i*4+3]
	}

	// Reset animation state
	g.offset = 0.0
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

	// Draw background particle field
	screenArea := g.width * g.height
	particleDensity := float64(screenArea) / 15000.0 // Adjust density for galaxy group scale

	aspectRatio := float64(g.width) / float64(g.height)
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
		orbitSpeed := g.particleOrbits[i] * g.orbitSpeed

		// More complex orbital motion
		baseAngle := math.Atan2(dy, dx)
		newAngle := baseAngle + orbitSpeed*(1.0-distFromCenter*0.3) // Faster orbits near center
		newRadius := math.Sqrt(dx*dx+dy*dy) * (1.0 + math.Sin(g.offset+phase)*0.1)

		// Calculate screen position with orbital motion
		screenX := int(float64(cx) + newRadius*math.Cos(newAngle)*float64(g.height)*aspectRatio*1.8)
		screenY := int(float64(cy) + newRadius*math.Sin(newAngle)*float64(g.height)*0.9)

		// Only draw if in bounds and not overlapping existing content
		if screenX >= 0 && screenX < g.width && screenY >= 0 && screenY < g.height && screen[screenY][screenX] == " " {
			particleType := g.particleTypes[i]

			var ch string
			if distFromCenter < 0.3 && particleType%5 == 0 {
				// Particles in the dense central region
				ch = g.majorStyle.Render("·")
			} else if particleType%7 == 0 {
				// Occasional brighter particles
				ch = g.satelliteStyle.Render("*")
			} else {
				// Background dust
				ch = g.dustStyle.Render("·")
			}
			screen[screenY][screenX] = ch
		}
	}

	// Update and draw major galaxies
	for i := range g.majorGalaxies {
		galaxy := &g.majorGalaxies[i]

		// Calculate galaxy center position with wider orbital paths
		aspectRatio := float64(g.width) / float64(g.height)
		galaxy.x = float64(cx) + math.Cos(galaxy.orbitAngle)*galaxy.orbitDist*float64(g.height)*aspectRatio*1.8
		galaxy.y = float64(cy) + math.Sin(galaxy.orbitAngle)*galaxy.orbitDist*float64(g.height)*0.9

		// Draw spiral arms with varying structure based on galaxy size
		arms := 4 + int(galaxy.size*2) // More arms for larger galaxies
		pointsPerArm := 35             // More points for denser arms

		// Calculate arm tightness based on galaxy size and screen dimensions
		screenScale := math.Min(float64(g.width), float64(g.height)) / 100.0
		armTightness := (0.3 + 0.15*math.Sin(galaxy.angle*0.5)) * screenScale

		for arm := 0; arm < arms; arm++ {
			// Use stored random values for arm variation
			armSeedBase := g.armSeeds[i][arm*2]
			armSeedTwist := g.armSeeds[i][arm*2+1]

			// Base angle with deterministic asymmetry
			armAngle := float64(arm)*2*math.Pi/float64(arms) + galaxy.angle +
				0.3*armSeedBase*math.Sin(float64(arm)+galaxy.angle)

			for p := 0; p < pointsPerArm; p++ {
				progress := float64(p) / float64(pointsPerArm)

				// Radius with enhanced arm length and perturbations
				baseRadius := math.Min(float64(g.width), float64(g.height)) / 4.2
				r := progress * galaxy.size * baseRadius

				// Add more complex spiral arm structure with stored randomness
				armWave := 0.2 * (0.8 + 0.4*armSeedBase) *
					math.Sin(progress*4+galaxy.angle) * (1.0 - progress*0.5)
				spiralTwist := math.Pow(progress, 0.6+0.2*armSeedTwist) // Variable winding
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
	}

	// Draw satellite galaxies
	for i := 0; i < g.satelliteCount; i++ {
		// Choose which major galaxy to orbit
		majorIndex := i % len(g.majorGalaxies)
		major := g.majorGalaxies[majorIndex]

		// Use stored random parameters for satellite orbits
		baseFreq := 1.0 + g.satelliteOrbits[i]*0.3
		satAngle := float64(i)*2*math.Pi/float64(g.satelliteCount) +
			g.offset*baseFreq +
			math.Sin(g.offset*0.5+g.satellitePhases[i])*0.2 // Deterministic perturbations

		// Enhanced satellite distribution with screen-aware scaling
		screenScale := math.Min(float64(g.width), float64(g.height)) / 6
		baseDist := screenScale * major.size

		// Use stored orbital parameters for distance variation
		satDist := baseDist * g.satelliteOrbits[i] * (0.8 +
			math.Sin(g.offset*0.7+g.satellitePhases[i])*0.2) // Dynamic variation

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
