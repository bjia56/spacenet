package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type StarCluster struct {
	*DefaultAnimation

	numStars      int     // Total number of stars
	clusterRadius float64 // Radius of the cluster
	nebulaRadius  float64 // Radius of the surrounding nebula
	turbulence    float64 // Amount of nebula movement
	brightness    float64 // Overall cluster brightness
	offset        float64 // Animation offset

	// Random parameters for deterministic rendering
	starSeeds     []float64 // Random values for star properties
	nebulaDensity []float64 // Random values for nebula density
	dustSeeds     []float64 // Random values for dust lane properties

	// Visual styles
	brightStarStyle lipgloss.Style // Bright main sequence stars
	faintStarStyle  lipgloss.Style // Fainter stars
	nebulaStyle     lipgloss.Style // Reflecting nebula
	dustStyle       lipgloss.Style // Dust lanes
}

type Star struct {
	x, y     float64
	bright   float64
	size     float64
	velocity float64
}

func NewStarCluster() *StarCluster {
	s := &StarCluster{
		numStars:      50,
		clusterRadius: 0.3,  // As fraction of height
		nebulaRadius:  0.45, // Larger than star cluster
		turbulence:    0.15,
		brightness:    0.8,
	}
	s.DefaultAnimation = NewDefaultAnimation(s)

	// Initialize styles with Pleiades-like colors
	s.brightStarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("159")) // Bright blue
	s.faintStarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("153"))  // Light blue
	s.nebulaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("24"))      // Dark blue
	s.dustStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))       // Dark gray

	return s
}

func (s *StarCluster) Tick() {
	s.offset += 0.03
}

func (s *StarCluster) ResetParameters() {
	// Initialize star parameters
	starBytes := s.RandBytes(s.numStars * 8)
	s.starSeeds = make([]float64, s.numStars)
	for i := range s.starSeeds {
		s.starSeeds[i] = float64(starBytes[i]) / 255.0
	}

	// Initialize nebula parameters
	nebulaBytes := s.RandBytes(200 * 8) // Match nebulaPoints in View
	s.nebulaDensity = make([]float64, 200)
	for i := range s.nebulaDensity {
		s.nebulaDensity[i] = float64(nebulaBytes[i]) / 255.0
	}

	// Initialize dust parameters
	dustBytes := s.RandBytes(30 * 8) // Match steps in View
	s.dustSeeds = make([]float64, 30)
	for i := range s.dustSeeds {
		s.dustSeeds[i] = float64(dustBytes[i]) / 255.0
	}

	// Reset animation offset
	s.offset = 0
}

func (s *StarCluster) View() string {
	screen := make([][]string, s.height)
	for i := range screen {
		screen[i] = make([]string, s.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	cx, cy := s.width/2, s.height/2

	// Draw the reflecting nebula first (background)
	nebulaPoints := 200
	for i := 0; i < nebulaPoints; i++ {
		angle := float64(i) * 2 * math.Pi / float64(nebulaPoints)
		// Add turbulent motion to nebula
		turbAngle := angle + math.Sin(s.offset+float64(i)*0.1)*s.turbulence

		// Vary the radius to create irregular shape
		radius := s.nebulaRadius * float64(s.height) *
			(1 + math.Sin(angle*3+s.offset)*0.2)

		x := int(float64(cx) + math.Cos(turbAngle)*radius*2)
		y := int(float64(cy) + math.Sin(turbAngle)*radius)

		if x >= 0 && x < s.width && y >= 0 && y < s.height {
			// Use pre-computed nebula density for deterministic patterns
			if s.nebulaDensity[i] > 0.3+math.Sin(angle*3+s.offset)*0.2 {
				screen[y][x] = s.nebulaStyle.Render("·")
			}
		}
	}

	// Draw dust lanes that intersect the nebula
	dustLines := 3
	for i := 0; i < dustLines; i++ {
		baseAngle := float64(i) * math.Pi / float64(dustLines)
		angle := baseAngle + math.Sin(s.offset*0.5)*0.3

		steps := 30
		for step := 0; step < steps; step++ {
			dist := float64(step) / float64(steps) * s.nebulaRadius * float64(s.height) * 2

			// Create wavy dust lanes
			wave := (s.dustSeeds[step]-0.5)*float64(s.height)*0.2 +
				math.Sin(float64(step)*0.3+s.offset)*float64(s.height)*0.1

			x := int(float64(cx) + math.Cos(angle)*dist)
			y := int(float64(cy) + math.Sin(angle)*dist*0.5 + wave)

			if x >= 0 && x < s.width && y >= 0 && y < s.height {
				screen[y][x] = s.dustStyle.Render("∙")
			}
		}
	}

	// Draw the stars
	for i := 0; i < s.numStars; i++ {
		// Calculate star position using pre-computed random values
		r := math.Pow(s.starSeeds[i], 0.5) * s.clusterRadius * float64(s.height)
		angle := float64(i)*math.Phi*2 + s.offset*math.Sin(s.starSeeds[i]*10)

		x := int(float64(cx) + math.Cos(angle)*r*2)
		y := int(float64(cy) + math.Sin(angle)*r)

		if x >= 0 && x < s.width && y >= 0 && y < s.height {
			// Vary star brightness based on position and time
			brightness := 0.5 + math.Sin(float64(i)+s.offset)*0.5

			// Choose star appearance based on brightness
			var ch string
			if i < 7 { // The Seven Sisters
				ch = s.brightStarStyle.Render("*")
			} else if brightness > 0.7 {
				ch = s.brightStarStyle.Render("+")
			} else {
				ch = s.faintStarStyle.Render("·")
			}
			screen[y][x] = ch

			// Add glowing effect around bright stars
			if i < 7 {
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx == 0 && dy == 0 {
							continue
						}
						gx, gy := x+dx, y+dy
						if gx >= 0 && gx < s.width && gy >= 0 && gy < s.height {
							if screen[gy][gx] == " " {
								screen[gy][gx] = s.nebulaStyle.Render("·")
							}
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
