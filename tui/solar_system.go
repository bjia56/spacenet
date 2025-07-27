package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type SolarSystem struct {
	*DefaultAnimation

	numPlanets int      // Number of planets
	planets    []Planet // Planet data
	offset     float64  // Animation offset

	// Visual styles
	starStyle     lipgloss.Style // Central star
	rockStyle     lipgloss.Style // Rocky planets
	gasStyle      lipgloss.Style // Gas giants
	iceStyle      lipgloss.Style // Ice giants
	moonStyle     lipgloss.Style // Moons
	ringStyle     lipgloss.Style // Planetary rings
	orbitStyle    lipgloss.Style // Orbit paths
	asteroidStyle lipgloss.Style // Asteroid belt
}

type Planet struct {
	size       float64    // Relative size
	orbitDist  float64    // Distance from star
	orbitSpeed float64    // Orbital velocity
	angle      float64    // Current orbital position
	moons      int        // Number of moons
	hasRings   bool       // Whether planet has rings
	ptype      PlanetType // Planet classification
}

type PlanetType int

const (
	Rocky PlanetType = iota
	Gas
	Ice
)

func NewSolarSystem() *SolarSystem {
	s := &SolarSystem{
		numPlanets: 8,
	}
	s.DefaultAnimation = NewDefaultAnimation(s)

	// Initialize planet data (roughly based on our solar system)
	s.planets = []Planet{
		{size: 0.4, orbitDist: 0.12, orbitSpeed: 4.15, ptype: Rocky},                           // Mercury
		{size: 0.9, orbitDist: 0.2, orbitSpeed: 1.62, ptype: Rocky},                            // Venus
		{size: 1.0, orbitDist: 0.28, orbitSpeed: 1.0, moons: 1, ptype: Rocky},                  // Earth
		{size: 0.5, orbitDist: 0.36, orbitSpeed: 0.53, moons: 2, ptype: Rocky},                 // Mars
		{size: 2.0, orbitDist: 0.52, orbitSpeed: 0.084, moons: 4, hasRings: true, ptype: Gas},  // Jupiter
		{size: 1.8, orbitDist: 0.65, orbitSpeed: 0.034, moons: 3, hasRings: true, ptype: Gas},  // Saturn
		{size: 1.2, orbitDist: 0.8, orbitSpeed: 0.012, moons: 2, hasRings: false, ptype: Ice},  // Uranus
		{size: 1.2, orbitDist: 0.92, orbitSpeed: 0.006, moons: 2, hasRings: false, ptype: Ice}, // Neptune
	}

	// Initialize styles
	s.starStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))     // Yellow
	s.rockStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("180"))     // Orange-brown
	s.gasStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("222"))      // Light yellow
	s.iceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("123"))      // Light blue
	s.moonStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))     // Light gray
	s.ringStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))     // Dark gray
	s.orbitStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))    // Very dark gray
	s.asteroidStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242")) // Medium gray

	return s
}

func (s *SolarSystem) Tick() {
	s.offset += 0.1
	// Update planet positions
	for i := range s.planets {
		s.planets[i].angle += s.planets[i].orbitSpeed * 0.02
	}
}

func (s *SolarSystem) View() string {
	screen := make([][]string, s.height)
	for i := range screen {
		screen[i] = make([]string, s.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	cx, cy := s.width/2, s.height/2

	// Draw orbit paths
	for i, planet := range s.planets {
		steps := 60
		for step := 0; step < steps; step++ {
			angle := float64(step) * 2 * math.Pi / float64(steps)
			x := int(float64(cx) + math.Cos(angle)*planet.orbitDist*float64(s.width)*0.8)
			y := int(float64(cy) + math.Sin(angle)*planet.orbitDist*float64(s.height)*0.7)

			if x >= 0 && x < s.width && y >= 0 && y < s.height {
				// Draw asteroid belt between 4th and 5th planets
				if i == 3 && step%3 == 0 {
					screen[y][x] = s.asteroidStyle.Render("·")
				} else if step%6 == 0 {
					screen[y][x] = s.orbitStyle.Render("·")
				}
			}
		}
	}

	// Draw the central star
	screen[cy][cx] = s.starStyle.Render("@")
	// Add star glow
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			x, y := cx+dx, cy+dy
			if x >= 0 && x < s.width && y >= 0 && y < s.height {
				screen[y][x] = s.starStyle.Render("·")
			}
		}
	}

	// Draw planets
	for _, planet := range s.planets {
		// Calculate planet position
		x := int(float64(cx) + math.Cos(planet.angle)*planet.orbitDist*float64(s.width)*0.8)
		y := int(float64(cy) + math.Sin(planet.angle)*planet.orbitDist*float64(s.height)*0.7)

		if x >= 0 && x < s.width && y >= 0 && y < s.height {
			// Choose planet appearance based on type
			var style lipgloss.Style
			var symbol string
			switch planet.ptype {
			case Rocky:
				style = s.rockStyle
				symbol = "o"
			case Gas:
				style = s.gasStyle
				symbol = "O"
			case Ice:
				style = s.iceStyle
				symbol = "0"
			}
			screen[y][x] = style.Render(symbol)

			// Draw rings if planet has them
			if planet.hasRings {
				ringRadius := 2
				for dx := -ringRadius; dx <= ringRadius; dx++ {
					rx := x + dx
					if rx >= 0 && rx < s.width && y >= 0 && y < s.height {
						if dx != 0 && screen[y][rx] == " " {
							screen[y][rx] = s.ringStyle.Render("-")
						}
					}
				}
			}

			// Draw moons
			for m := 0; m < planet.moons; m++ {
				moonAngle := planet.angle + float64(m)*2.5 + s.offset
				moonDist := float64(planet.size + 1.5)
				mx := int(float64(x) + math.Cos(moonAngle)*moonDist)
				my := int(float64(y) + math.Sin(moonAngle)*moonDist*0.5)

				if mx >= 0 && mx < s.width && my >= 0 && my < s.height {
					screen[my][mx] = s.moonStyle.Render(".")
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
