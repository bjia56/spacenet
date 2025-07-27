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
		numFilaments:  5,   // Increased number of filaments
		pointsPerFil:  100, // More points for denser appearance
		wallCurvature: 0.4, // Increased curve for more vertical spread
		flowSpeed:     0.05,
		wallLength:    30, // Increased length for wider spread
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
		// Calculate filament offset with improved distribution
		filOffset := (float64(fil) - float64(w.numFilaments)/2) * 2.5 / float64(w.numFilaments)
		for p := 0; p < w.pointsPerFil; p++ {
			// Calculate position along the curved wall with spread adjustment
			progress := float64(p) / float64(w.pointsPerFil)
			spreadFactor := 1.0 + math.Sin(progress*math.Pi)*0.2 // Wider in the middle
			x := progress * w.wallLength * spreadFactor

			// Complex wave effect combining multiple frequencies
			phase := x*(0.2+0.05*float64(fil)) + math.Sin(w.offset*0.3)*0.2

			// Primary wave combines three frequencies for organic movement
			wave := math.Sin(x*0.5+w.offset+phase)*0.6 +
				math.Sin(x*0.25+w.offset*1.5)*0.3 +
				math.Sin(x*0.75+w.offset*0.7)*0.1

			// Add distance-based damping for more natural flow
			damping := 1.0 - math.Pow(math.Abs(progress-0.5)*2, 2)*0.3
			wave *= damping

			// Create curved structure with varying height
			y := wave * w.wallCurvature * float64(w.height/4)

			// Dynamic vertical offset with time-varying component
			timeShift := math.Sin(w.offset*0.2+float64(fil)) * 0.3
			y += filOffset*float64(w.height/6)*(1.0+timeShift) +
				math.Sin(phase+w.offset*0.4)*float64(w.height/12)

			// Scale and position in screen space with improved screen utilization
			screenX := int(float64(cx) + (x-w.wallLength/2)*float64(w.width)/20) // Scale based on screen width
			screenY := int(float64(cy) + y*float64(w.height)/12)                 // Scale based on screen height

			// Draw main filament point
			if screenX >= 0 && screenX < w.width && screenY >= 0 && screenY < w.height {
				// Vary the appearance based on position and screen location
				var ch string
				distFromCenter := math.Sqrt(math.Pow(float64(screenX-cx), 2) + math.Pow(float64(screenY-cy), 2))
				if p%7 == 0 || distFromCenter < float64(w.height)/6 {
					ch = w.clusterStyle.Render("*") // Galaxy clusters
				} else if p%3 == 0 {
					ch = w.galaxyStyle.Render(".") // Individual galaxies
				} else {
					ch = w.filamentStyle.Render("·") // Filament matter
				}
				screen[screenY][screenX] = ch

				// Dynamic branch generation
				branchProb := 0.15 + math.Sin(progress*math.Pi*2+w.offset)*0.05
				if starRand(int(x*1000)) < branchProb {
					// Calculate branch length based on position and flow
					flowStrength := math.Abs(wave) + math.Abs(math.Sin(w.offset+phase))
					branchLen := 3.0 + math.Sin(phase*2+w.offset)*2.0 - flowStrength

					// Create two branches in opposite directions
					for b := 1.0; b <= branchLen; b++ {
						// Branch angle varies with position, time, and flow
						baseAngle := math.Atan2(y-float64(cy), x-float64(cx))
						branchAngle := baseAngle +
							math.Sin(phase+w.offset*0.5)*math.Pi/3 +
							math.Sin(w.offset*0.7+float64(fil))*math.Pi/6

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
