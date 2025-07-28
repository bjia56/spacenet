package main

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

type StarCluster struct {
	*DefaultAnimation

	numStars      int     // Total number of stars in the cluster
	centerStars   int     // Number of bright central stars
	offset        float64 // Animation offset
	rotationSpeed float64 // Speed of rotation

	// Pre-calculated star parameters
	starPositions [][2]float64 // Positions relative to center
	starColors    []lipgloss.Style
	starTypes     []rune    // Different star symbols
	starPhases    []float64 // Individual star movement phases
	pulsePeriods  []float64 // Different pulse periods for stars

	// Visual styles for different star types
	blueStyle   lipgloss.Style // Hot blue stars
	whiteStyle  lipgloss.Style // White stars
	yellowStyle lipgloss.Style // Yellow stars
	redStyle    lipgloss.Style // Red stars

	// Background particle styles (more subdued versions)
	particleBlueStyle   lipgloss.Style // Dim blue particles
	particleWhiteStyle  lipgloss.Style // Dim white particles
	particleYellowStyle lipgloss.Style // Dim yellow particles
	particleRedStyle    lipgloss.Style // Dim red particles

	// Background particle field parameters
	particlePositions [][2]float64     // Normalized positions (0-1)
	particlePhases    []float64        // Phase offsets for particle movement
	particleOrbits    []float64        // Orbital parameters for particles
	particleTypes     []byte           // Type of each particle (for visual variation)
	particleColors    []lipgloss.Style // Individual particle colors
}

func NewStarCluster() *StarCluster {
	j := &StarCluster{
		numStars:      40, // Total stars in visualization
		centerStars:   7,  // Prominent central stars
		offset:        0.0,
		rotationSpeed: 0.02,
	}
	j.DefaultAnimation = NewDefaultAnimation(j)
	return j
}

func (j *StarCluster) Tick() {
	j.offset += j.rotationSpeed
}

func (j *StarCluster) ResetParameters() {
	// Get random bytes for initialization
	randParams := j.RandBytes(8)
	colorParams := j.RandBytes(8) // Additional random bytes for color variation

	// Adjust parameters with some randomness
	j.numStars = 35 + int(randParams[0]%15)  // 35-49 stars
	j.centerStars = 6 + int(randParams[1]%3) // 6-8 central stars
	j.rotationSpeed = 0.02 + float64(randParams[2]%20)/1000.0

	// Initialize color styles with random variations
	// Star colors (bright and bold)
	blueBase := 51 + int(colorParams[0]%3) - 1    // 50-53 range
	whiteBase := 15 + int(colorParams[1]%3) - 1   // 14-17 range
	yellowBase := 220 + int(colorParams[2]%3) - 1 // 219-222 range
	redBase := 196 + int(colorParams[3]%3) - 1    // 195-198 range

	j.blueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", blueBase))).Bold(true)
	j.whiteStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", whiteBase))).Bold(true)
	j.yellowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", yellowBase))).Bold(true)
	j.redStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", redBase))).Bold(true)

	// Background particle colors (subdued, not bold)
	particleBlueBase := 17 + int(colorParams[4]%3) - 1    // 16-19 range
	particleWhiteBase := 242 + int(colorParams[5]%3) - 1  // 241-244 range
	particleYellowBase := 136 + int(colorParams[6]%3) - 1 // 135-138 range
	particleRedBase := 131 + int(colorParams[7]%3) - 1    // 130-133 range

	j.particleBlueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", particleBlueBase)))
	j.particleWhiteStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", particleWhiteBase)))
	j.particleYellowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", particleYellowBase)))
	j.particleRedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fmt.Sprintf("%d", particleRedBase)))

	// Initialize star parameters
	j.starPositions = make([][2]float64, j.numStars)
	j.starColors = make([]lipgloss.Style, j.numStars)
	j.starTypes = make([]rune, j.numStars)
	j.starPhases = make([]float64, j.numStars)
	j.pulsePeriods = make([]float64, j.numStars)

	// Generate star parameters
	starBytes := j.RandBytes(j.numStars * 4) // 4 bytes per star

	// First place the central bright stars
	for i := 0; i < j.centerStars; i++ {
		angle := (float64(i) * 2 * math.Pi / float64(j.centerStars)) +
			float64(starBytes[i])/255.0*0.5 // Small random offset

		// Central stars are arranged in a small circle
		dist := 2.0 + float64(starBytes[i*4+1])/255.0*2.0
		j.starPositions[i] = [2]float64{
			math.Cos(angle) * dist,
			math.Sin(angle) * dist,
		}

		// Assign colors to central stars - mix of bright stars
		colorRoll := starBytes[i*4+2] % 4
		switch colorRoll {
		case 0:
			j.starColors[i] = j.blueStyle
			j.starTypes[i] = '*'
		case 1:
			j.starColors[i] = j.whiteStyle
			j.starTypes[i] = '*'
		case 2:
			j.starColors[i] = j.yellowStyle
			j.starTypes[i] = '*'
		case 3:
			j.starColors[i] = j.redStyle
			j.starTypes[i] = '*'
		}

		j.starPhases[i] = float64(starBytes[i*4+3]) / 255.0 * math.Pi * 2
		j.pulsePeriods[i] = 0.5 + float64(starBytes[i*4+3])/255.0 // 0.5-1.5 range
	}

	// Place remaining stars in a wider cluster
	for i := j.centerStars; i < j.numStars; i++ {
		// Create a spiral-like distribution
		angle := float64(i) * math.Pi * (1 + math.Sqrt(5))
		dist := math.Sqrt(float64(i-j.centerStars))*2.0 +
			float64(starBytes[i*4])/255.0*3.0

		j.starPositions[i] = [2]float64{
			math.Cos(angle) * dist,
			math.Sin(angle) * dist,
		}

		// Assign varied appearances to outer stars
		colorRoll := starBytes[i*4+1] % 8
		switch colorRoll {
		case 0, 1:
			j.starColors[i] = j.blueStyle
			j.starTypes[i] = '·'
		case 2, 3:
			j.starColors[i] = j.whiteStyle
			j.starTypes[i] = '·'
		case 4, 5:
			j.starColors[i] = j.yellowStyle
			j.starTypes[i] = '·'
		case 6, 7:
			j.starColors[i] = j.redStyle
			j.starTypes[i] = '·'
		}

		j.starPhases[i] = float64(starBytes[i*4+2]) / 255.0 * math.Pi * 2
		j.pulsePeriods[i] = 0.3 + float64(starBytes[i*4+3])/255.0 // 0.3-1.3 range
	}

	j.offset = 0.0

	// Initialize background particle field with fixed number of particles
	const baseParticles = 240 // Base number of particles for star cluster background
	j.particlePositions = make([][2]float64, baseParticles)
	j.particlePhases = make([]float64, baseParticles)
	j.particleOrbits = make([]float64, baseParticles)
	j.particleTypes = make([]byte, baseParticles)
	j.particleColors = make([]lipgloss.Style, baseParticles)

	// Get random bytes for particle initialization
	particleBytes := j.RandBytes(baseParticles * 5) // 5 bytes per particle (x, y, orbit, type, color)
	for i := 0; i < baseParticles; i++ {
		// Convert to polar coordinates for better distribution
		angle := float64(particleBytes[i*5]) / 255.0 * math.Pi * 2
		radius := 0.1 + float64(particleBytes[i*5+1])/255.0*0.9 // Radial distribution

		// Store in normalized coordinates (0-1 range)
		j.particlePositions[i][0] = 0.5 + math.Cos(angle)*radius*0.5
		j.particlePositions[i][1] = 0.5 + math.Sin(angle)*radius*0.5

		// Phase and orbit parameters
		j.particlePhases[i] = float64(particleBytes[i*5+2]) / 255.0 * math.Pi * 2
		j.particleOrbits[i] = 0.3 + float64(particleBytes[i*5+3])/255.0*0.7 // 0.3-1.0 range

		// Assign particle colors based on position and random variation
		colorRoll := particleBytes[i*5+4] % 8
		switch colorRoll {
		case 0:
			j.particleColors[i] = j.particleBlueStyle
		case 1, 2:
			j.particleColors[i] = j.particleWhiteStyle
		case 3, 4, 5:
			j.particleColors[i] = j.particleYellowStyle
		default:
			j.particleColors[i] = j.particleRedStyle
		}
		j.particleTypes[i] = particleBytes[i*5+4]
	}
}

func (j *StarCluster) View() string {
	// Create empty screen
	screen := make([][]string, j.height)
	for i := range screen {
		screen[i] = make([]string, j.width)
		for k := range screen[i] {
			screen[i][k] = " "
		}
	}

	// Calculate center of screen
	cx, cy := j.width/2, j.height/2

	// Draw background particle field
	screenArea := j.width * j.height
	particleDensity := float64(screenArea) / 12000.0 // Adjust density for star cluster scale

	for i, pos := range j.particlePositions {
		// Skip some particles based on screen size to maintain consistent density
		if float64(i) > float64(len(j.particlePositions))*particleDensity {
			break
		}

		// Calculate radial distance from center
		dx := pos[0] - 0.5
		dy := pos[1] - 0.5
		distFromCenter := math.Sqrt(dx*dx+dy*dy) * 2.0

		// Calculate orbital motion
		phase := j.particlePhases[i]
		orbitSpeed := j.particleOrbits[i] * j.rotationSpeed * 0.3

		// Complex orbital motion
		baseAngle := math.Atan2(dy, dx)
		newAngle := baseAngle + orbitSpeed*(1.0-distFromCenter*0.4) // Faster orbits near center
		newRadius := math.Sqrt(dx*dx+dy*dy) * (1.0 + math.Sin(j.offset*0.5+phase)*0.15)

		// Calculate screen position with aspect ratio correction
		screenX := int(float64(cx) + newRadius*math.Cos(newAngle)*float64(j.width))
		screenY := int(float64(cy) + newRadius*math.Sin(newAngle)*float64(j.height))

		// Only draw if in bounds and not overlapping
		if screenX >= 0 && screenX < j.width && screenY >= 0 && screenY < j.height && screen[screenY][screenX] == " " {
			particleType := j.particleTypes[i]
			var ch string

			if distFromCenter < 0.3 && particleType%5 == 0 {
				// Dense central region particles
				ch = j.particleColors[i].Render("·")
			} else if particleType%7 == 0 {
				// Occasional brighter particles
				ch = j.particleColors[i].Render("*")
			} else {
				// Background particles
				ch = j.particleColors[i].Render("∙")
			}
			screen[screenY][screenX] = ch
		}
	}

	// Draw all stars
	for i := 0; i < j.numStars; i++ {
		// Calculate rotation and pulsation
		rotAngle := j.offset * (1.0 - float64(i)/float64(j.numStars)*0.5)
		cos, sin := math.Cos(rotAngle), math.Sin(rotAngle)

		// Rotate position
		x := j.starPositions[i][0]*cos - j.starPositions[i][1]*sin
		y := j.starPositions[i][0]*sin + j.starPositions[i][1]*cos

		// Add subtle movement
		phase := j.starPhases[i]
		period := j.pulsePeriods[i]
		moveX := math.Sin(j.offset*period+phase) * 0.5
		moveY := math.Cos(j.offset*period+phase) * 0.5

		// Scale and position on screen, using both dimensions
		baseScale := math.Min(float64(j.width), float64(j.height)) / 25.0
		aspectRatio := float64(j.width) / float64(j.height)
		scaleX := baseScale * math.Min(aspectRatio, 1.5) // Allow some horizontal stretch, but not too much
		scaleY := baseScale

		screenX := int(x*scaleX + float64(cx) + moveX)
		screenY := int(y*scaleY + float64(cy) + moveY)

		// Draw star if in bounds
		if screenX >= 0 && screenX < j.width && screenY >= 0 && screenY < j.height {
			// Calculate brightness based on position and time
			brightness := math.Sin(j.offset*period+phase)*0.2 + 0.8

			// Vary appearance based on brightness
			var ch string
			if i < j.centerStars {
				// Central stars maintain their bright appearance
				ch = j.starColors[i].Render(string(j.starTypes[i]))
			} else if brightness > 0.9 {
				// Occasional bright pulses for outer stars
				ch = j.starColors[i].Render("*")
			} else {
				ch = j.starColors[i].Render(string(j.starTypes[i]))
			}

			screen[screenY][screenX] = ch
		}
	}

	// Build output string
	var output string
	for _, row := range screen {
		for _, ch := range row {
			output += ch
		}
		output += "\n"
	}
	return output
}
