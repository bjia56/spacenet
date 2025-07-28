package main

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

type Supercluster struct {
	*DefaultAnimation

	numClusters   int     // Number of major galaxy clusters
	pointsPerArm  int     // Points per filamentary arm
	numArms       int     // Number of spiral arms converging on Great Attractor
	flowSpeed     float64 // Animation speed
	attractorPull float64 // Strength of Great Attractor influence
	offset        float64 // Animation offset

	// Stored random parameters
	clusterOffsets []float64 // Random offset for each cluster
	flowSeeds      []byte    // Random seeds for flow patterns
	clusterSizes   []float64 // Size variation for clusters

	// Visual styles
	galaxyStyle    lipgloss.Style // Individual galaxies
	clusterStyle   lipgloss.Style // Dense galaxy clusters
	attractorStyle lipgloss.Style // Great Attractor
	filamentStyle  lipgloss.Style // Connecting filaments
}

func NewSupercluster() *Supercluster {
	l := &Supercluster{
		numClusters:   12,   // Major galaxy clusters
		pointsPerArm:  150,  // Density of filaments
		numArms:       8,    // Number of major filamentary arms
		flowSpeed:     0.03, // Base flow speed
		attractorPull: 0.4,  // Strength of gravitational influence
		offset:        0.0,

		// Initialize styles with cosmic-appropriate colors
		galaxyStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true),  // Deep blue-white for galaxies
		clusterStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("159")),            // Bright cyan for clusters
		attractorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true), // Warm pink-red for attractor
		filamentStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("60")),             // Deep purple-blue for filaments
	}
	l.DefaultAnimation = NewDefaultAnimation(l)
	return l
}

func (l *Supercluster) Tick() {
	l.offset += l.flowSpeed
}

func (l *Supercluster) ResetParameters() {
	randParams := l.RandBytes(8)

	// Randomize structure parameters
	l.numClusters = 10 + int(randParams[0]%6)    // 10-15 clusters
	l.pointsPerArm = 150 + int(randParams[1]%50) // 150-199 points
	l.numArms = 7 + int(randParams[2]%5)         // 7-11 arms
	l.flowSpeed = 0.03 + float64(randParams[3]%20)/1000.0
	l.attractorPull = 0.4 + float64(randParams[4]%30)/100.0

	// Initialize cluster parameters
	l.clusterOffsets = make([]float64, l.numClusters)
	l.clusterSizes = make([]float64, l.numClusters)
	offsetBytes := l.RandBytes(l.numClusters * 2)
	for i := range l.clusterOffsets {
		l.clusterOffsets[i] = float64(offsetBytes[i])/255.0*0.5 - 0.25
		l.clusterSizes[i] = 0.5 + float64(offsetBytes[i+l.numClusters])/255.0
	}

	// Initialize flow pattern seeds
	l.flowSeeds = l.RandBytes(l.pointsPerArm * l.numArms)

	// Randomize colors while maintaining cosmic theme
	colorBytes := l.RandBytes(4)

	// Galaxy colors - variations of blue-white (base: 39)
	galaxyColor := 38 + (colorBytes[0] % 3) // Range 38-40
	l.galaxyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", galaxyColor))).Bold(true)

	// Cluster colors - variations of bright cyan (base: 159)
	clusterColor := 158 + (colorBytes[1] % 3) // Range 158-160
	l.clusterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", clusterColor)))

	// Attractor colors - variations of warm reds/pinks (base: 203)
	attractorColor := 202 + (colorBytes[2] % 3) // Range 202-204
	l.attractorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", attractorColor))).Bold(true)

	// Filament colors - variations of deep purple-blue (base: 60)
	filamentColor := 59 + (colorBytes[3] % 3) // Range 59-61
	l.filamentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", filamentColor)))

	// Reset animation state
	l.offset = 0.0
}

func (l *Supercluster) View() string {
	// Create empty screen
	screen := make([][]string, l.height)
	for i := range screen {
		screen[i] = make([]string, l.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	cx, cy := l.width/2, l.height/2

	// Draw the Great Attractor
	gaSize := 2
	for dy := -gaSize; dy <= gaSize; dy++ {
		for dx := -gaSize; dx <= gaSize; dx++ {
			x, y := cx+dx, cy+dy
			if x >= 0 && x < l.width && y >= 0 && y < l.height {
				if dx*dx+dy*dy <= gaSize*gaSize {
					screen[y][x] = l.attractorStyle.Render("@")
				}
			}
		}
	}

	// Draw filamentary arms
	for arm := 0; arm < l.numArms; arm++ {
		baseAngle := float64(arm) * 2 * math.Pi / float64(l.numArms)
		for p := 0; p < l.pointsPerArm; p++ {
			progress := float64(p) / float64(l.pointsPerArm)

			// Spiral arm path with inward flow
			radius := float64(l.width/3) * (1 - math.Pow(progress, 0.7))
			angle := baseAngle +
				progress*1.5 + // Spiral twist
				math.Sin(progress*math.Pi*2+l.offset)*0.2 // Flow movement

			// Add flow towards Great Attractor
			flowStrength := l.attractorPull * progress
			radius *= 1 - flowStrength*math.Sin(l.offset)

			// Calculate position
			x := float64(cx) + radius*math.Cos(angle)
			y := float64(cy) + radius*math.Sin(angle)

			// Add turbulence
			seedIndex := arm*l.pointsPerArm + p
			turbulence := float64(l.flowSeeds[seedIndex])/255.0 - 0.5
			x += turbulence * 4
			y += turbulence * 4

			screenX, screenY := int(x), int(y)

			if screenX >= 0 && screenX < l.width && screenY >= 0 && screenY < l.height {
				// Vary point appearance based on position and flow
				var ch string
				if p%8 == 0 {
					ch = l.clusterStyle.Render("*")
					// Add cluster density
					if l.width > 100 {
						for dy := -1; dy <= 1; dy++ {
							for dx := -1; dx <= 1; dx++ {
								nx, ny := screenX+dx, screenY+dy
								if nx >= 0 && nx < l.width && ny >= 0 && ny < l.height &&
									screen[ny][nx] == " " {
									screen[ny][nx] = l.galaxyStyle.Render("·")
								}
							}
						}
					}
				} else if p%3 == 0 {
					ch = l.galaxyStyle.Render("·")
				} else {
					ch = l.filamentStyle.Render("·")
				}
				screen[screenY][screenX] = ch
			}
		}
	}

	// Add major galaxy clusters
	for i := 0; i < l.numClusters; i++ {
		angle := float64(i) * 2 * math.Pi / float64(l.numClusters)
		radius := float64(l.width/4) * (0.6 + l.clusterOffsets[i])

		// Orbital movement around Great Attractor
		angle += l.offset * (0.5 + l.clusterOffsets[i]*0.3)

		x := float64(cx) + radius*math.Cos(angle)
		y := float64(cy) + radius*math.Sin(angle)

		screenX, screenY := int(x), int(y)

		if screenX >= 0 && screenX < l.width && screenY >= 0 && screenY < l.height {
			// Draw cluster with size variation
			clusterSize := int(3 * l.clusterSizes[i])
			for dy := -clusterSize; dy <= clusterSize; dy++ {
				for dx := -clusterSize; dx <= clusterSize; dx++ {
					if dx*dx+dy*dy <= clusterSize*clusterSize {
						nx, ny := screenX+dx, screenY+dy
						if nx >= 0 && nx < l.width && ny >= 0 && ny < l.height {
							screen[ny][nx] = l.clusterStyle.Render("*")
						}
					}
				}
			}
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
