package main

import (
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
}

func NewStarCluster() *StarCluster {
	j := &StarCluster{
		numStars:      40, // Total stars in visualization
		centerStars:   7,  // Prominent central stars
		offset:        0.0,
		rotationSpeed: 0.02,

		// Initialize styles for different stellar types
		blueStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true),  // Bright blue
		whiteStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true),  // Bright white
		yellowStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true), // Golden yellow
		redStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true), // Bright red
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

	// Adjust parameters with some randomness
	j.numStars = 35 + int(randParams[0]%15)  // 35-49 stars
	j.centerStars = 6 + int(randParams[1]%3) // 6-8 central stars
	j.rotationSpeed = 0.02 + float64(randParams[2]%20)/1000.0

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
