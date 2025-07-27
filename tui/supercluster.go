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
		numClusters:       5,
		pointsPerCluster:  60,
		flowSpeed:         0.03,
		attractorStrength: 0.8,
		offset:            0.0,

		// Initialize styles for different components
		galaxyStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("45")),  // Cyan
		clusterStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("87")),  // Light cyan
		attractorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
		flowStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("25")),  // Dark blue
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
		// Calculate cluster center position
		baseAngle := float64(cluster)*2*math.Pi/float64(s.numClusters) + s.rotationAngle
		clusterDist := float64(s.height) * 0.3 // Base distance from center

		// Add some variety to cluster distances
		clusterDist += math.Sin(float64(cluster)*1.5) * float64(s.height) * 0.1

		basex := float64(cx) + math.Cos(baseAngle)*clusterDist
		basey := float64(cy) + math.Sin(baseAngle)*clusterDist*0.5 // Elliptical shape

		// Draw galaxies in and around each cluster
		for p := 0; p < s.pointsPerCluster; p++ {
			// Calculate galaxy position with flow towards attractor
			progress := float64(p) / float64(s.pointsPerCluster)
			angle := baseAngle + math.Sin(progress*math.Pi*2+s.offset)*0.5

			// Distance varies with time to create flowing effect
			dist := clusterDist * (1.0 - progress*s.attractorStrength*
				(0.5+0.5*math.Sin(progress*5+s.offset)))

			// Calculate final position with flowing motion
			x := int(basex + math.Cos(angle)*dist*0.3)
			y := int(basey + math.Sin(angle)*dist*0.15)

			if x >= 0 && x < s.width && y >= 0 && y < s.height {
				// Vary the appearance based on position and flow
				var ch string
				distToCenter := math.Sqrt(float64((x-cx)*(x-cx) + (y-cy)*(y-cy)))

				if p < s.pointsPerCluster/8 && distToCenter > float64(s.height)/4 {
					ch = s.clusterStyle.Render("*") // Cluster centers
				} else if p%3 == 0 {
					ch = s.galaxyStyle.Render(".") // Individual galaxies
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
