package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type PlanetView struct {
	*DefaultAnimation

	planetType    PlanetKind // Type of planet
	rotationSpeed float64    // Speed of planetary rotation
	numMoons      int        // Number of moons
	hasRings      bool       // Whether planet has rings
	ringWidth     float64    // Width of ring system
	offset        float64    // Animation offset
	atmosphere    bool       // Whether planet has visible atmosphere

	// Surface features
	bands        int     // Number of atmospheric/surface bands
	bandSpeed    float64 // Speed of band rotation
	spotSize     float64 // Size of major spot/storm (if gas giant)
	spotLat      float64 // Latitude of the spot
	spotRotation float64 // Current rotation angle of spot

	// Visual styles
	surfaceStyle    lipgloss.Style // Main planet surface
	spotStyle       lipgloss.Style // Special features (like storms)
	bandStyle       lipgloss.Style // Atmospheric bands
	ringStyle       lipgloss.Style // Planetary rings
	moonStyle       lipgloss.Style // Moons
	atmosphereStyle lipgloss.Style // Atmosphere glow
	shadowStyle     lipgloss.Style // Shadow effects
}

type PlanetKind int

const (
	GasGiant PlanetKind = iota
	RockyPlanet
	IceGiant
)

func NewPlanet(kind PlanetKind) *PlanetView {
	p := &PlanetView{
		planetType:    kind,
		rotationSpeed: 0.05,
		offset:        0,
	}
	p.DefaultAnimation = NewDefaultAnimation(p)

	// Set characteristics based on planet type
	switch kind {
	case GasGiant:
		p.bands = 5
		p.bandSpeed = 0.03
		p.spotSize = 0.15
		p.spotLat = 0.3
		p.numMoons = 4
		p.hasRings = true
		p.ringWidth = 0.8
		p.atmosphere = true

		// Jupiter-like colors
		p.surfaceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("222"))    // Light yellow
		p.spotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("160"))       // Red
		p.bandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("215"))       // Orange
		p.ringStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))       // Gray
		p.atmosphereStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("230")) // Light yellow

	case RockyPlanet:
		p.bands = 3
		p.bandSpeed = 0.01
		p.numMoons = 1
		p.hasRings = false
		p.atmosphere = true

		// Earth-like colors
		p.surfaceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("32"))     // Green
		p.spotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))        // Blue
		p.bandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("195"))       // Light blue
		p.atmosphereStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("195")) // Light blue

	case IceGiant:
		p.bands = 4
		p.bandSpeed = 0.02
		p.numMoons = 2
		p.hasRings = true
		p.ringWidth = 0.4
		p.atmosphere = true

		// Neptune-like colors
		p.surfaceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))     // Cyan
		p.spotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))        // Blue
		p.bandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("45"))        // Light blue
		p.ringStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))       // Gray
		p.atmosphereStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("195")) // Light cyan
	}

	p.moonStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))   // Light gray
	p.shadowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236")) // Dark gray

	return p
}

func (p *PlanetView) Tick() {
	p.offset += p.rotationSpeed
	p.spotRotation += p.bandSpeed
}

func (p *PlanetView) View() string {
	screen := make([][]string, p.height)
	for i := range screen {
		screen[i] = make([]string, p.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	cx, cy := p.width/2, p.height/2
	radius := float64(p.height) * 0.35

	// Draw rings behind planet if present
	if p.hasRings {
		p.drawRings(screen, cx, cy, radius, true)
	}

	// Draw the planet's atmosphere glow
	if p.atmosphere {
		glowRadius := radius * 1.2
		for angle := 0.0; angle < 2*math.Pi; angle += 0.1 {
			x := int(float64(cx) + math.Cos(angle)*glowRadius*2)
			y := int(float64(cy) + math.Sin(angle)*glowRadius)

			if x >= 0 && x < p.width && y >= 0 && y < p.height {
				screen[y][x] = p.atmosphereStyle.Render("·")
			}
		}
	}

	// Draw the planet surface with bands
	for y := cy - int(radius); y <= cy+int(radius); y++ {
		// Calculate width of this row
		dy := float64(y - cy)
		dx := math.Sqrt(radius*radius - dy*dy)

		// Scale dx for elliptical appearance
		dx *= 2 // Stretch horizontally

		for x := cx - int(dx); x <= cx+int(dx); x++ {
			if x >= 0 && x < p.width && y >= 0 && y < p.height {
				// Calculate surface coordinates
				lat := math.Asin(dy/radius) / (math.Pi / 2)
				long := math.Atan2(float64(x-cx), dx) + p.offset

				// Determine surface features
				var ch string
				bandIndex := int((lat + 1) * float64(p.bands/2))

				// Check if we're in the spot area (for gas giants)
				inSpot := false
				if p.planetType == GasGiant {
					spotLong := math.Mod(long+p.spotRotation, 2*math.Pi)
					spotLat := lat - p.spotLat
					if math.Abs(spotLat) < p.spotSize &&
						math.Abs(math.Sin(spotLong)) < p.spotSize*2 {
						inSpot = true
					}
				}

				if inSpot {
					ch = p.spotStyle.Render("@")
				} else if bandIndex%2 == 0 {
					ch = p.surfaceStyle.Render("o")
				} else {
					ch = p.bandStyle.Render("O")
				}
				screen[y][x] = ch
			}
		}
	}

	// Draw rings in front of planet
	if p.hasRings {
		p.drawRings(screen, cx, cy, radius, false)
	}

	// Draw moons
	for m := 0; m < p.numMoons; m++ {
		moonAngle := float64(m)*2.5 + p.offset*0.5
		moonDist := radius * 1.8
		mx := int(float64(cx) + math.Cos(moonAngle)*moonDist*2)
		my := int(float64(cy) + math.Sin(moonAngle)*moonDist)

		if mx >= 0 && mx < p.width && my >= 0 && my < p.height {
			screen[my][mx] = p.moonStyle.Render("o")

			// Add shadow effect for moons
			shadowX := mx + 1
			if shadowX >= 0 && shadowX < p.width {
				screen[my][shadowX] = p.shadowStyle.Render("·")
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

func (p *PlanetView) drawRings(screen [][]string, cx, cy int, planetRadius float64, behind bool) {
	ringRadius := planetRadius * 2.2
	for angle := 0.0; angle < 2*math.Pi; angle += 0.05 {
		// Calculate elliptical ring coordinates
		x := int(float64(cx) + math.Cos(angle)*ringRadius*2)
		y := int(float64(cy) + math.Sin(angle)*ringRadius*0.3)

		if x >= 0 && x < p.width && y >= 0 && y < p.height {
			// Only draw rings that should be visible based on position
			if behind && y < cy || !behind && y >= cy {
				if screen[y][x] == " " {
					// Vary ring density based on position
					if int(angle*10)%2 == 0 {
						screen[y][x] = p.ringStyle.Render("-")
					} else {
						screen[y][x] = p.ringStyle.Render("·")
					}
				}
			}
		}
	}
}
