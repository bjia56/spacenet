package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type GreatWall struct {
	*DefaultAnimation

	numFilaments  int     // Number of filament strands
	pointsPerFil  int     // Points per filament
	wallCurvature float64 // How much the wall curves
	flowSpeed     float64 // Animation speed
	wallLength    float64 // Length of the wall structure
	offset        float64 // Animation offset

	// Visual styles for different parts of the wall
	galaxyStyle   lipgloss.Style // Individual galaxy points
	clusterStyle  lipgloss.Style // Dense galaxy clusters
	filamentStyle lipgloss.Style // Connecting filaments
}

func NewGreatWall() *GreatWall {
	w := &GreatWall{
		numFilaments:  3,
		pointsPerFil:  80,
		wallCurvature: 0.3,
		flowSpeed:     0.05,
		wallLength:    20,
		offset:        0.0,

		// Initialize lipgloss styles for the great wall
		galaxyStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true), // Light blue
		clusterStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("51")),            // Bright cyan
		filamentStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("27")),            // Deep blue
	}
	w.DefaultAnimation = NewDefaultAnimation(w)
	return w
}

func (w *GreatWall) Tick() {
	w.offset += w.flowSpeed
}

func (w *GreatWall) View() string {
	// Create empty screen
	screen := make([][]string, w.height)
	for i := range screen {
		screen[i] = make([]string, w.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	cx, cy := w.width/2, w.height/2

	// Draw multiple filaments that form the wall
	for fil := 0; fil < w.numFilaments; fil++ {
		filOffset := float64(fil) * 2.0 / float64(w.numFilaments)
		for p := 0; p < w.pointsPerFil; p++ {
			// Calculate position along the curved wall
			progress := float64(p) / float64(w.pointsPerFil)
			x := progress * w.wallLength

			// Add flowing wave effect with rotation
			phase := x * (0.2 + 0.05*float64(fil)) // This creates the rotation effect
			wave := math.Sin(x*0.5 + w.offset + phase)

			// Create curved structure with varying height
			y := wave * w.wallCurvature * float64(w.height/4)

			// Add vertical offset for multiple filaments with slight wave variation
			y += filOffset*float64(w.height/6) + math.Sin(phase)*float64(w.height/12)

			// Scale and position in screen space
			screenX := int(float64(cx) + (x-w.wallLength/2)*1.5)
			screenY := int(float64(cy) + y)

			// Draw main filament point
			if screenX >= 0 && screenX < w.width && screenY >= 0 && screenY < w.height {
				// Vary the appearance based on position
				var ch string
				if p%7 == 0 {
					ch = w.clusterStyle.Render("*") // Galaxy clusters
				} else if p%3 == 0 {
					ch = w.galaxyStyle.Render(".") // Individual galaxies
				} else {
					ch = w.filamentStyle.Render("·") // Filament matter
				}
				screen[screenY][screenX] = ch

				// Add branches at regular intervals
				if p%10 == 0 {
					// Calculate branch length based on position
					branchLen := 4.0 - (math.Abs(wave) * 2.0) // Shorter branches at peaks

					// Create two branches in opposite directions
					for b := 1.0; b <= branchLen; b++ {
						// Branch angle varies with position and time
						branchAngle := math.Sin(phase+w.offset*0.5) * math.Pi / 3

						// Calculate branch endpoints
						dx1 := int(b * math.Cos(branchAngle))
						dy1 := int(b * math.Sin(branchAngle))
						dx2 := int(b * math.Cos(branchAngle+math.Pi))
						dy2 := int(b * math.Sin(branchAngle+math.Pi))

						// Draw branch points if in bounds
						bx1, by1 := screenX+dx1, screenY+dy1
						if bx1 >= 0 && bx1 < w.width && by1 >= 0 && by1 < w.height {
							screen[by1][bx1] = w.filamentStyle.Render("·")
						}

						bx2, by2 := screenX+dx2, screenY+dy2
						if bx2 >= 0 && bx2 < w.width && by2 >= 0 && by2 < w.height {
							screen[by2][bx2] = w.filamentStyle.Render("·")
						}
					}
				}
			}
		}
	}

	// Build the output string
	var output string
	for _, row := range screen {
		for _, ch := range row {
			output += ch
		}
		output += "\n"
	}
	return output
}
