package main

import (
	"fmt"
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

	// Stored random parameters for deterministic rendering
	clusterDistances []float64 // Random distance variations for each cluster
	clusterPhases    []float64 // Random phase shifts for each cluster
	filamentWidths   []float64 // Random width variations for each cluster
	branchPatterns   []float64 // Random branching patterns
	turbulenceSeeds  []float64 // Random seeds for turbulence

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

func (s *Supercluster) ResetParameters() {
	// Get random bytes for initial parameters
	randParams := s.RandBytes(8)

	// Set core parameters with random variations
	s.numClusters = 10 + int(randParams[0]%6)        // 10-15 clusters
	s.pointsPerCluster = 100 + int(randParams[1]%80) // 100-179 points
	s.flowSpeed = 0.015 + float64(randParams[2]%20)/1000.0
	s.attractorStrength = 0.75 + float64(randParams[3]%30)/100.0

	// Reset animation state
	s.offset = 0.0
	s.rotationAngle = float64(randParams[4]) / 255.0 * math.Pi // Random initial rotation

	// Initialize arrays for stored random parameters
	s.clusterDistances = make([]float64, s.numClusters)
	s.clusterPhases = make([]float64, s.numClusters)
	s.filamentWidths = make([]float64, s.numClusters)
	s.branchPatterns = make([]float64, s.pointsPerCluster)
	s.turbulenceSeeds = make([]float64, s.pointsPerCluster)

	// Get random bytes for all parameters
	distBytes := s.RandBytes(s.numClusters)
	phaseBytes := s.RandBytes(s.numClusters)
	widthBytes := s.RandBytes(s.numClusters)
	branchBytes := s.RandBytes(s.pointsPerCluster)
	turbBytes := s.RandBytes(s.pointsPerCluster)

	// Initialize cluster-specific random parameters
	for i := 0; i < s.numClusters; i++ {
		s.clusterDistances[i] = float64(distBytes[i]) / 255.0 * 0.4   // 0.0-0.4 variation
		s.clusterPhases[i] = float64(phaseBytes[i]) / 255.0 * math.Pi // 0-π phase shift
		s.filamentWidths[i] = 0.3 + float64(widthBytes[i])/255.0*0.4  // 0.3-0.7 width
	}

	// Initialize point-specific random parameters
	for i := 0; i < s.pointsPerCluster; i++ {
		s.branchPatterns[i] = float64(branchBytes[i]) / 255.0      // 0.0-1.0
		s.turbulenceSeeds[i] = float64(turbBytes[i]) / 255.0 * 0.2 // 0.0-0.2
	}

	// Randomize colors
	colorBytes := s.RandBytes(8)

	// Galaxy colors - variations of bright blue-cyan (base: 51)
	galaxyColor := 45 + (colorBytes[0] % 7) // Range 45-51 (bright blue to cyan)
	s.galaxyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", galaxyColor))).Bold(true)

	// Cluster colors - variations of light cyan (base: 159)
	clusterColor := 153 + (colorBytes[1] % 7) // Range 153-159 (cyan variations)
	s.clusterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", clusterColor))).Bold(true)

	// Attractor colors - variations of magenta-red (base: 201)
	attractorColor := 196 + (colorBytes[2] % 6) // Range 196-201 (red to magenta)
	s.attractorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", attractorColor))).Bold(true)

	// Flow colors - variations of dark blue (base: 17)
	flowColor := 17 + (colorBytes[3] % 6) // Range 17-22 (dark blue variations)
	s.flowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", flowColor)))
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

		// Use stored random parameters for this cluster
		filamentPhase := s.clusterPhases[cluster]
		distVariation := s.clusterDistances[cluster]

		// Calculate varying distances using stored random parameters
		clusterDist := baseClusterDist * (1.0 +
			distVariation + // Base variation
			math.Sin(s.offset*0.2+filamentPhase)*0.15) // Time-based motion

		// Create asymmetric distribution using stored parameters
		ellipticalFactor := 1.8 + s.filamentWidths[cluster]*0.4 // Varying width
		verticalFactor := 0.6 + math.Cos(filamentPhase)*0.2     // Varying height

		// Calculate base position with sheet-like structure
		basex := float64(cx) + math.Cos(baseAngle)*clusterDist*ellipticalFactor
		basey := float64(cy) + math.Sin(baseAngle)*clusterDist*verticalFactor

		// Draw galaxies in and around each cluster
		for p := 0; p < s.pointsPerCluster; p++ {
			// Calculate galaxy position with filamentary flow patterns
			progress := float64(p) / float64(s.pointsPerCluster)

			// Use stored random parameters for branching and structure
			branchPattern := s.branchPatterns[p]
			turbSeed := s.turbulenceSeeds[p]
			filamentWidth := s.filamentWidths[cluster] * (1.0 - 0.5*progress) // Thinner towards ends

			// Complex angular variation using stored parameters
			angle := baseAngle +
				branchPattern*filamentWidth + // Fixed branch pattern
				math.Sin(progress*math.Pi*3+s.offset)*0.3*filamentWidth + // Main flow
				math.Sin(progress*5+s.offset*0.5)*0.15*filamentWidth // Fine structure

			// Enhanced distance calculation with stored turbulence
			flowStrength := s.attractorStrength * (1.0 + 0.2*math.Sin(progress*6+s.offset))
			flowEffect := math.Pow(progress, 1.2) * flowStrength
			turbulence := turbSeed * math.Sin(s.offset*1.5) * (1.0 - progress) // Controlled turbulence
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
