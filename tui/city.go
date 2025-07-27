package main

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

type City struct {
	*DefaultAnimation

	numBuildings   int     // Number of buildings
	numVehicles    int     // Number of vehicles in traffic
	dayNightCycle  float64 // Day/night cycle progress
	weatherEffect  float64 // Current weather intensity
	trafficDensity float64 // Density of traffic flows
	offset         float64 // Animation offset

	// City characteristics
	buildings []Building
	roads     []Road

	// Visual styles
	buildingStyle   lipgloss.Style // Main building structure
	windowStyle     lipgloss.Style // Building windows
	roadStyle       lipgloss.Style // Roads and highways
	vehicleStyle    lipgloss.Style // Moving vehicles
	skylineStyle    lipgloss.Style // Background buildings
	beaconStyle     lipgloss.Style // Warning beacons
	highlightStyle  lipgloss.Style // Lit windows
	backgroundStyle lipgloss.Style // Sky color
}

type Building struct {
	x, y      int     // Position
	height    int     // Height of building
	width     int     // Width of building
	style     int     // Architectural style
	lit       float64 // Percentage of windows lit
	hasBeacon bool    // Whether building has warning beacon
}

type Road struct {
	startX, startY int     // Start position
	endX, endY     int     // End position
	traffic        float64 // Traffic density
}

func NewCity() *City {
	c := &City{
		numBuildings:   15,
		numVehicles:    30,
		trafficDensity: 0.7,
	}
	c.DefaultAnimation = NewDefaultAnimation(c)

	// Initialize styles for different times of day
	c.updateStyles(0) // Start at night

	return c
}

func (c *City) Tick() {
	c.offset += 0.02
	c.dayNightCycle = (math.Sin(c.offset*0.1) + 1) * 0.5 // Cycle between 0 (night) and 1 (day)
	c.weatherEffect = math.Sin(c.offset * 0.3)           // Varies between -1 and 1

	// Update styles based on time of day
	c.updateStyles(c.dayNightCycle)
}

func (c *City) updateStyles(timeOfDay float64) {
	// Calculate style colors based on time of day
	if timeOfDay < 0.3 { // Night
		c.buildingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))  // Dark gray
		c.windowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))    // Yellow
		c.roadStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("234"))      // Very dark gray
		c.vehicleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))   // Red
		c.skylineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))   // Dark background
		c.beaconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))    // Red
		c.highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Bright yellow
		c.backgroundStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("17")) // Dark blue
	} else if timeOfDay < 0.7 { // Day
		c.buildingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))   // Light gray
		c.windowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("195"))     // Light blue
		c.roadStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))       // Medium gray
		c.vehicleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))    // Gray
		c.skylineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("251"))    // Light background
		c.beaconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))     // White
		c.highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))  // White
		c.backgroundStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("153")) // Light blue
	} else { // Sunset/Sunrise
		c.buildingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))   // Medium gray
		c.windowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))     // Orange
		c.roadStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))       // Dark gray
		c.vehicleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))    // Orange-red
		c.skylineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))    // Medium background
		c.beaconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))     // Red
		c.highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))  // Orange
		c.backgroundStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203")) // Pink-orange
	}
}

func (c *City) View() string {
	if len(c.buildings) == 0 {
		c.generateCityLayout()
	}

	screen := make([][]string, c.height)
	for i := range screen {
		screen[i] = make([]string, c.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	// Draw sky/background with optional weather effects
	c.drawBackground(screen)

	// Draw distant skyline
	c.drawSkyline(screen)

	// Draw buildings
	c.drawBuildings(screen)

	// Draw roads and traffic
	c.drawRoads(screen)

	var output string
	for _, row := range screen {
		for _, ch := range row {
			output += ch
		}
		output += "\n"
	}
	return output
}

func (c *City) generateCityLayout() {
	// Generate main buildings
	c.buildings = make([]Building, c.numBuildings)
	for i := range c.buildings {
		c.buildings[i] = Building{
			x:         int(rand(i*3) * float64(c.width)),
			width:     3 + int(rand(i*7)*4),
			height:    5 + int(rand(i*11)*float64(c.height)/2),
			style:     int(rand(i*13) * 3),
			lit:       rand(i * 17),
			hasBeacon: rand(i*19) > 0.7,
		}
		c.buildings[i].y = c.height - c.buildings[i].height
	}

	// Generate road network
	baseRoads := 3
	c.roads = make([]Road, baseRoads)
	for i := range c.roads {
		y := c.height - 2 - i*3
		c.roads[i] = Road{
			startX:  0,
			startY:  y,
			endX:    c.width,
			endY:    y,
			traffic: 0.5 + rand(i*23)*0.5,
		}
	}
}

func (c *City) drawBackground(screen [][]string) {
	// Draw sky with weather effects
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			if rand(x*y) > 0.9 {
				if c.weatherEffect > 0.5 { // Rain
					screen[y][x] = c.backgroundStyle.Render("|")
				} else if c.weatherEffect < -0.5 { // Snow
					screen[y][x] = c.backgroundStyle.Render("*")
				}
			}
		}
	}
}

func (c *City) drawSkyline(screen [][]string) {
	// Draw distant buildings in the background
	skylineHeight := c.height / 3
	for x := 0; x < c.width; x++ {
		height := int(math.Sin(float64(x)*0.2+c.offset)*float64(skylineHeight/4)) + skylineHeight
		for y := c.height - height; y < c.height; y++ {
			if y >= 0 && screen[y][x] == " " {
				screen[y][x] = c.skylineStyle.Render("░")
			}
		}
	}
}

func (c *City) drawBuildings(screen [][]string) {
	for _, building := range c.buildings {
		// Draw building structure
		for y := building.y; y < c.height; y++ {
			for x := building.x; x < building.x+building.width && x < c.width; x++ {
				if x >= 0 && y >= 0 && x < c.width && y < c.height {
					// Draw windows
					if (x-building.x)%(building.style+2) == 1 && (y-building.y)%3 == 1 {
						if rand(int(c.offset)*x*y) < building.lit {
							screen[y][x] = c.highlightStyle.Render("■")
						} else {
							screen[y][x] = c.windowStyle.Render("□")
						}
					} else {
						screen[y][x] = c.buildingStyle.Render("█")
					}
				}
			}
		}

		// Add building beacon
		if building.hasBeacon && building.y > 0 {
			beaconPhase := math.Sin(c.offset*2) > 0
			if beaconPhase {
				screen[int(math.Min(float64(building.y-1), float64(c.height-1)))][int(math.Min(float64(building.x+building.width/2), float64(c.width-1)))] = c.beaconStyle.Render("*")
			}
		}
	}
}

func (c *City) drawRoads(screen [][]string) {
	for _, road := range c.roads {
		// Draw road
		for x := road.startX; x < road.endX; x++ {
			y := road.startY
			if y >= 0 && y < c.height && x >= 0 && x < c.width {
				screen[y][x] = c.roadStyle.Render("═")
			}
		}

		// Draw traffic
		vehicleCount := int(float64(c.width) * road.traffic * c.trafficDensity)
		for i := 0; i < vehicleCount; i++ {
			x := int(float64(c.width) * (float64(i)/float64(vehicleCount) + math.Sin(c.offset+float64(i))*0.1))
			if x >= 0 && x < c.width {
				if i%2 == 0 {
					screen[road.startY][x] = c.vehicleStyle.Render("►")
				} else {
					screen[road.startY][x] = c.vehicleStyle.Render("◄")
				}
			}
		}
	}
}

// Simple deterministic random number generator for consistent layouts
func cityRand(seed int) float64 {
	x := float64(seed * 12345)
	return (math.Sin(x) + 1.0) * 0.5
}
