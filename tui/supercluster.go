package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type Supercluster struct {
	*DefaultAnimation

	numClusters       int     // Number of major galaxy clusters
	pointsPerCluster  int     // Points representing galaxies per cluster
	flowSpeed         float64 // Animation speed
	attractorStrength float64 // Strength of the great attractor
	offset            float64 // Animation offset
	rotationAngle     float64 // Overall rotation of the structure

	// Visual styles for different components
	galaxyStyle    lipgloss.Style // Individual galaxies
	clusterStyle   lipgloss.Style // Dense cluster centers
	attractorStyle lipgloss.Style // Great attractor
	flowStyle      lipgloss.Style // Flow lines
}

func NewSupercluster() *Supercluster {
	s := &Supercluster{
		numClusters:       12,   // More filaments for Laniakea-like structure
		pointsPerCluster:  120,  // More points for denser filaments
		flowSpeed:         0.02, // Slower for more stable flow
		attractorStrength: 0.85, // Stronger pull towards center
		offset:            0.0,

		// Initialize styles for different components with cosmic colors
		galaxyStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true),  // Bright blue-cyan
		clusterStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("159")).Bold(true), // Light cyan
		attractorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true), // Magenta-red for Virgo Supercluster
		flowStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("17")),             // Dark blue for cosmic web
	}
	s.DefaultAnimation = NewDefaultAnimation(s)
	return s
}

func (s *Supercluster) Tick() {
	s.offset += s.flowSpeed
	s.rotationAngle += s.flowSpeed * 0.1
}

func (s *Supercluster) View() string {
	screen := make([][]string, s.height)
	for i := range screen {
		screen[i] = make([]string, s.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	cx, cy := s.width/2, s.height/2

	// Draw the great attractor at the center
	screen[cy][cx] = s.attractorStyle.Render("@")

	// Draw each major cluster
	for cluster := 0; cluster < s.numClusters; cluster++ {
		// Calculate filament base position with Laniakea-like distribution
		baseAngle := float64(cluster)*2*math.Pi/float64(s.numClusters) + s.rotationAngle

		// Create varying distances for different filaments
		baseClusterDist := math.Min(float64(s.width), float64(s.height)*2) * 0.4

		// Each filament has a unique flow pattern
		filamentPhase := float64(cluster) * math.Pi / 6

		// Calculate varying distances to create sheets and filaments
		clusterDist := baseClusterDist * (1.0 +
			math.Sin(filamentPhase)*0.3 + // Sheet-like structure
			math.Sin(float64(cluster)*2.7)*0.2 + // Length variation
			math.Sin(s.offset*0.2+float64(cluster))*0.15) // Time-based motion

		// Create asymmetric distribution like Laniakea
		ellipticalFactor := 1.8 + math.Sin(baseAngle*2)*0.4 // Varying width
		verticalFactor := 0.6 + math.Cos(baseAngle*3)*0.2   // Varying height

		// Calculate base position with sheet-like structure
		basex := float64(cx) + math.Cos(baseAngle)*clusterDist*ellipticalFactor
		basey := float64(cy) + math.Sin(baseAngle)*clusterDist*verticalFactor

		// Draw galaxies in and around each cluster
		for p := 0; p < s.pointsPerCluster; p++ {
			// Calculate galaxy position with filamentary flow patterns
			progress := float64(p) / float64(s.pointsPerCluster)

			// Create branching filaments
			branchPhase := math.Sin(progress*6 + float64(cluster)) // Branch variation
			filamentWidth := 0.4 + 0.3*math.Pow(progress, 2.0)     // Filaments get thinner

			// Complex angular variation for filamentary structure
			angle := baseAngle +
				branchPhase*filamentWidth + // Branch spread
				math.Sin(progress*math.Pi*3+s.offset)*0.3*filamentWidth + // Main flow
				math.Sin(progress*7+s.offset*0.5)*0.15*filamentWidth // Fine structure

			// Enhanced distance calculation for flowing filaments
			flowStrength := s.attractorStrength * (1.0 + 0.2*math.Sin(progress*8+s.offset))
			flowEffect := math.Pow(progress, 1.2) * flowStrength                      // Non-linear flow
			turbulence := math.Sin(progress*12+s.offset*1.5) * 0.1 * (1.0 - progress) // Detail
			dist := clusterDist * (1.0 - flowEffect + turbulence)

			// Calculate final position with filamentary structure
			spread := 0.2 + 0.2*math.Pow(1.0-progress, 2.0) // Wider at ends
			x := int(basex + math.Cos(angle)*dist*spread)
			y := int(basey + math.Sin(angle)*dist*spread)

			if x >= 0 && x < s.width && y >= 0 && y < s.height {
				// Vary the appearance based on position and flow
				var ch string
				distToCenter := math.Sqrt(float64((x-cx)*(x-cx) + (y-cy)*(y-cy)))

				// Calculate brightness based on position and cluster density
				brightness := 1.0 - (distToCenter / float64(s.height))
				brightness = math.Max(0.2, math.Min(1.0, brightness*1.5))

				if p < s.pointsPerCluster/8 && distToCenter > float64(s.height)/4 {
					ch = s.clusterStyle.Render("@") // Bright cluster centers
				} else if p%3 == 0 {
					if brightness > 0.7 {
						ch = s.galaxyStyle.Render("●") // Bright galaxies
					} else {
						ch = s.galaxyStyle.Render("∘") // Dimmer galaxies
					}
				} else {
					ch = s.flowStyle.Render("·") // Flow lines
				}
				screen[y][x] = ch
			}

			// Draw flow lines towards attractor
			if p%5 == 0 && progress > 0.3 {
				// Calculate points along flow line
				flowx := float64(x)
				flowy := float64(y)
				for f := 0; f < 3; f++ {
					// Move towards attractor
					flowx = flowx*0.8 + float64(cx)*0.2
					flowy = flowy*0.8 + float64(cy)*0.2
					fx, fy := int(flowx), int(flowy)

					if fx >= 0 && fx < s.width && fy >= 0 && fy < s.height {
						if screen[fy][fx] == " " {
							screen[fy][fx] = s.flowStyle.Render("∙")
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
