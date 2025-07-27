package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/bjia56/gosendip"
	"github.com/bjia56/spacenet/server/api"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle         = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	statusMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	errorMessageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	tableStyle         = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
	helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

// Tables
type level int

const (
	t16 level = iota
	t32
	t48
	t64
	t80
	t96
	t112
	t128
)

type UnitTables [8]table.Model
type UnitColumns [3]table.Column

func (ut *UnitTables) Initialize() {
	columns := UnitColumns{
		{Title: "Subnet", Width: 50},
		{Title: "Owner", Width: 30},
		{Title: "Percentage", Width: 20},
	}
	for i := range ut {
		ut[i] = table.New(
			table.WithColumns([]table.Column(columns[:])),
			table.WithRows([]table.Row{}),
			table.WithFocused(true),
			table.WithHeight(10),
		)
	}
}

func (ut *UnitTables) SetHeight(height int) {
	for i := range ut {
		ut[i].SetHeight(height)
	}
}

func (ut *UnitTables) SetWidth(width int) {
	for i := range ut {
		ut[i].SetWidth(width)
		columns := ut[i].Columns()
		columns[0].Width = (width * 5) / 10
		columns[1].Width = (width * 3) / 10
		columns[2].Width = width - (columns[0].Width + columns[1].Width) - 6
		ut[i].SetColumns(columns)
	}
}

// Animations
type UnitAnimations [8]Animation

func (ua *UnitAnimations) Initialize() {
	*ua = UnitAnimations{
		NewGreatWall(),
		NewSupercluster(),
		NewGalaxyGroup(),
		NewGalaxy(),
		NewStarCluster(),
		NewSolarSystem(),
		NewPlanet(),
		NewCity(),
	}
}

func (ua *UnitAnimations) SetDimensions(width, height int) {
	for i := range ua {
		ua[i].SetDimensions(width, height)
	}
}

func (ua *UnitAnimations) Ticker() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timer.TickMsg{ID: int(uintptr(unsafe.Pointer(ua)))}
	})
}

func (ua *UnitAnimations) Update(msg tea.Msg, level level) tea.Cmd {
	switch msg := msg.(type) {
	case timer.TickMsg:
		if msg.ID == int(uintptr(unsafe.Pointer(ua))) {
			ua[level].Tick()
			return ua.Ticker()
		}
	}
	return nil
}

// Block granularity mappings
var subnetMappings = [8]int{
	t16:  16,
	t32:  32,
	t48:  48,
	t64:  64,
	t80:  80,
	t96:  96,
	t112: 112,
	t128: 128,
}

// Model represents the state of our application
type Model struct {
	serverAddr string
	httpPort   int
	udpPort    int
	name       string

	unitTables    UnitTables // Tables for displaying subnets with fun names
	shadowTables  UnitTables // For shadowing the current table with actual IPv6 addresses
	selections    [8]string  // Selected subnets for each table level
	viewing       level
	refreshClaims bool // Whether to refresh claims on the next update

	animationModels UnitAnimations // Animation models for visualizations

	statusMessage string
	errorMessage  string
}

func makeIPv6Full(i int, prefix string, level level) (string, int) {
	makeFull := func() string {
		hex := fmt.Sprintf("%04x", i)
		numSubBlocks := 8 - (int(level) + 1)
		zeroBlocks := strings.Repeat(":0000", numSubBlocks)
		if prefix == "" {
			return fmt.Sprintf("%s%s", hex, zeroBlocks)
		}
		return fmt.Sprintf("%s%s%s", prefix, hex, zeroBlocks)
	}
	full := makeFull()
	return full, subnetMappings[level]
}

// Initialize returns an initial model
func Initialize(serverAddr string, httpPort, udpPort int, name string) *Model {
	m := &Model{
		serverAddr:    serverAddr,
		httpPort:      httpPort,
		udpPort:       udpPort,
		name:          name,
		refreshClaims: true,
	}
	m.unitTables.Initialize()
	m.shadowTables.Initialize()
	m.PopulateTable("", t16)
	m.animationModels.Initialize()
	return m
}

// SendClaim sends a claim for an IP
func (m *Model) SendClaim(ip string) (string, error) {
	// Build sendip args
	args := []string{"-d", m.name, "-p", "ipv6", "-6s", ip, "-p", "udp", "-ud", fmt.Sprintf("%d", m.udpPort), m.serverAddr}

	// Execute sendip command
	rc, _ := gosendip.SendIP(args)
	if rc != 0 {
		return "", fmt.Errorf("failed to send claim for %s: exit code %d", ip, rc)
	}

	return "Claim sent successfully!", nil
}

// PopulateTable populates a table with 2^16 rows
func (m *Model) PopulateTable(prefix string, level level) {
	rows := make([]table.Row, 0, 1<<16)
	shadowRows := make([]table.Row, 0, 1<<16)
	for i := range 1 << 16 {
		addr, subnet := makeIPv6Full(i, prefix, level)
		name, err := GenerateName(addr, subnet)
		if err != nil {
			panic(fmt.Sprintf("Failed to generate name for %s: %v", addr, err))
		}
		rows = append(rows, table.Row{
			name,
			"", // Placeholder for owner,
			"", // Placeholder for percentage
		})
		shadowRows = append(shadowRows, table.Row{
			fmt.Sprintf("%s/%d", addr, subnet),
		})
	}
	m.unitTables[level].SetRows(rows)
	m.shadowTables[level].SetRows(shadowRows)
}

// FetchClaims fetches claims for a range of subnets
func (m *Model) FetchClaims(prefix string, level level, start, end int) {
	for i := max(start, 0); i < min(end, 1<<16); i++ {
		addr, subnet := makeIPv6Full(i, prefix, level)
		serverUrl := fmt.Sprintf("http://%s:%d/api/subnet/%s/%d", m.serverAddr, m.httpPort, addr, subnet)

		client := &http.Client{}
		req, err := http.NewRequest("GET", serverUrl, nil)
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error fetching claims: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error fetching claims: %s %v", serverUrl, resp.StatusCode)
			return
		}

		// Process the response
		subnetResp := &api.SubnetResponse{}
		if err := json.NewDecoder(resp.Body).Decode(subnetResp); err != nil {
			log.Printf("Error decoding response: %v", err)
			return
		}

		// Update the table with the claim
		row := m.unitTables[level].Rows()[i]
		row[1] = subnetResp.Owner
		if subnetResp.Percentage > 0 {
			row[2] = strconv.FormatFloat(subnetResp.Percentage, 'f', 2, 64) + "%"
		}
		m.unitTables[level].SetRows(m.unitTables[level].Rows())
	}
}

// GetParentSelection returns the parent selection for a given level
func (m *Model) GetParentSelection(level level) string {
	if level == t16 {
		return ""
	}
	parentLevel := level - 1
	return m.selections[parentLevel]
}

// Init initializes the application
func (m *Model) Init() tea.Cmd {
	m.animationModels[m.viewing].AnimateForIP(net.IPv6zero)
	return tea.Batch(
		m.animationModels.Ticker(),
	)
}

// Update handles user input and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		reserved := 6
		m.unitTables.SetHeight(msg.Height - reserved)
		m.unitTables.SetWidth((msg.Width / 2) - 2)
		m.animationModels.SetDimensions(m.unitTables[0].Width(), m.unitTables[0].Height())

	case tea.KeyMsg:
		m.statusMessage = ""
		m.errorMessage = ""

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			if m.viewing > 0 {
				m.viewing--
				m.refreshClaims = true
			}

		case "enter":
			cursor := m.unitTables[m.viewing].Cursor()
			selection := m.shadowTables[m.viewing].Rows()[cursor][0]
			if m.viewing < t128 {
				selection = selection[:5*(m.viewing+1)] // Adjust for the level
				m.selections[m.viewing] = selection
				m.viewing++
				m.PopulateTable(m.selections[m.viewing-1], m.viewing)
			} else {
				// At the last level, send a claim
				ip := strings.Split(selection, "/")[0] // Get the IP part before the prefix
				if msg, err := m.SendClaim(ip); err == nil {
					m.statusMessage = statusMessageStyle.Render(msg)
					m.errorMessage = ""
				} else {
					m.errorMessage = errorMessageStyle.Render("Failed to send claim")
					m.statusMessage = ""
				}
			}
			m.refreshClaims = true
		}
	}

	// Update the selected row in the current table
	lastCursor := m.unitTables[m.viewing].Cursor()
	t, cmd := m.unitTables[m.viewing].Update(msg)
	m.unitTables[m.viewing] = t
	cmds = append(cmds, cmd)
	newCursor := m.unitTables[m.viewing].Cursor()
	if lastCursor != newCursor {
		m.refreshClaims = true // Refresh claims if cursor changed
	}

	ip := strings.Split(m.shadowTables[m.viewing].Rows()[newCursor][0], "/")[0]
	ipv6 := net.ParseIP(ip)
	m.animationModels[m.viewing].AnimateForIP(ipv6)

	// Update the animation model
	if cmd := m.animationModels.Update(msg, m.viewing); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current state of the model
func (m *Model) View() string {
	if m.refreshClaims {
		activeTable := m.unitTables[m.viewing]
		m.FetchClaims(m.GetParentSelection(m.viewing), m.viewing, activeTable.Cursor()-activeTable.Height(), activeTable.Cursor()+activeTable.Height())
		m.refreshClaims = false
	}

	msg := m.statusMessage
	if m.errorMessage != "" {
		msg = m.errorMessage
	}

	return titleStyle.Render("SpaceNet Browser") + "\n\n" +
		lipgloss.JoinHorizontal(
			lipgloss.Bottom,
			tableStyle.Render(m.unitTables[m.viewing].View()),
			tableStyle.Render(m.animationModels[m.viewing].View()),
		) + "\n" + msg + "\n" +
		helpStyle("enter: select subnet, esc: back, q: quit")
}

func main() {
	// Parse command line flags
	server := flag.String("server", "::1", "IPv6 address of the server")
	httpPort := flag.Int("http-port", 8080, "HTTP port for the server's API")
	udpPort := flag.Int("port", 1337, "UDP port number of the server")
	name := flag.String("name", "Anonymous", "Name to use for claims")
	flag.Parse()

	// Set up logging
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("Fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	// Initialize the TUI
	p := tea.NewProgram(Initialize(*server, *httpPort, *udpPort, *name), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
