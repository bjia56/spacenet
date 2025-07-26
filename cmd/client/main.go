// A TUI client for the SpaceNet server
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bjia56/gosendip"
	"github.com/bjia56/spacenet/api"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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

func (ut *UnitTables) Focus() {
	for i := range ut {
		ut[i].Focus()
	}
}

func (ut *UnitTables) Blur() {
	for i := range ut {
		ut[i].Blur()
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
	}
}

// Block granularity mappings
var unitMappings = [8]string{
	t16:  "Superstructure",
	t32:  "Supercluster",
	t48:  "Galaxy Group",
	t64:  "Galaxy",
	t80:  "Star Group",
	t96:  "Solar System",
	t112: "Planet",
	t128: "City",
}
var subnetMappings = [8]string{
	t16:  "16",
	t32:  "32",
	t48:  "48",
	t64:  "64",
	t80:  "80",
	t96:  "96",
	t112: "112",
	t128: "128",
}

// Model represents the state of our application
type Model struct {
	serverAddr string
	httpPort   int
	udpPort    int
	name       string

	nameInput     textinput.Model
	inputSelected bool
	unitTables    UnitTables
	tableSelected bool
	selections    [8]string // Selected subnets for each table level
	viewing       level

	statusMessage string
	errorMessage  string
}

func makeIPv6Full(i int, prefix string, level level) string {
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
	return full + "/" + subnetMappings[level]
}

// Initialize returns an initial model
func Initialize(serverAddr string, httpPort, udpPort int) *Model {
	columns := []table.Column{
		{Title: "Subnet", Width: 50},
		{Title: "Owner", Width: 30},
	}

	ti := textinput.New()
	ti.Prompt = "Claim for: "
	ti.SetValue("Anonymous")
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 33

	m := &Model{
		serverAddr:    serverAddr,
		httpPort:      httpPort,
		udpPort:       udpPort,
		name:          "Anonymous",
		nameInput:     ti,
		inputSelected: true,
		unitTables: UnitTables{
			t16: table.New(
				table.WithColumns(columns),
				table.WithRows([]table.Row{}),
				table.WithFocused(false),
				table.WithHeight(10),
			),
			t32: table.New(
				table.WithColumns(columns),
				table.WithRows([]table.Row{}),
				table.WithFocused(false),
				table.WithHeight(10),
			),
			t48: table.New(
				table.WithColumns(columns),
				table.WithRows([]table.Row{}),
				table.WithFocused(false),
				table.WithHeight(10),
			),
			t64: table.New(
				table.WithColumns(columns),
				table.WithRows([]table.Row{}),
				table.WithFocused(false),
				table.WithHeight(10),
			),
			t80: table.New(
				table.WithColumns(columns),
				table.WithRows([]table.Row{}),
				table.WithFocused(false),
				table.WithHeight(10),
			),
			t96: table.New(
				table.WithColumns(columns),
				table.WithRows([]table.Row{}),
				table.WithFocused(false),
				table.WithHeight(10),
			),
			t112: table.New(
				table.WithColumns(columns),
				table.WithRows([]table.Row{}),
				table.WithFocused(false),
				table.WithHeight(10),
			),
			t128: table.New(
				table.WithColumns(columns),
				table.WithRows([]table.Row{}),
				table.WithFocused(false),
				table.WithHeight(10),
			),
		},
		tableSelected: false,
	}
	m.PopulateTable("", t16)
	return m
}

// SendClaim sends a claim for an IP
func (m *Model) SendClaim(ip string) tea.Cmd {
	return func() tea.Msg {
		// Build sendip args
		args := []string{"-d", m.name, "-p", "ipv6", "-6s", ip, "-p", "udp", "-ud", fmt.Sprintf("%d", m.udpPort), m.serverAddr}

		// Execute sendip command
		rc, _ := gosendip.SendIP(args)
		if rc != 0 {
			return fmt.Errorf("failed to send claim for %s: exit code %d", ip, rc)
		}

		return "Claim sent successfully!"
	}
}

// PopulateTable populates a table with 2^16 rows
func (m *Model) PopulateTable(prefix string, level level) {
	rows := make([]table.Row, 0, 1<<16)
	for i := range 1 << 16 {
		row := table.Row{
			makeIPv6Full(i, prefix, level),
			"", // Placeholder for owner
		}
		rows = append(rows, row)
	}
	m.unitTables[level].SetRows(rows)
}

// FetchClaims fetches claims for a range of subnets
func (m *Model) FetchClaims(prefix string, level level, start, end int) {
	for i := max(start, 0); i < min(end, 1<<16); i++ {
		shorthand := makeIPv6Full(i, prefix, level)
		serverUrl := fmt.Sprintf("http://%s:%d/api/subnet/%s", m.serverAddr, m.httpPort, shorthand)

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
	return textinput.Blink
}

// Update handles user input and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		reserved := 10
		m.unitTables.SetHeight(msg.Height - reserved)
		m.unitTables.SetWidth(msg.Width - 2)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if m.tableSelected {
				return m, tea.Quit
			} else {
				break
			}

		case "esc":
			if !m.tableSelected {
				break
			}

			if m.viewing > 0 {
				m.viewing--
			}

		case "enter":
			if !m.tableSelected {
				m.tableSelected = true
				m.inputSelected = false
				m.nameInput.Blur()
				m.unitTables.Focus()
			} else {
				selection := m.unitTables[m.viewing].SelectedRow()[0]
				selection = selection[:5*(m.viewing+1)] // Adjust for the level
				m.selections[m.viewing] = selection
				if m.viewing < t128 {
					m.viewing++
					m.PopulateTable(m.selections[m.viewing-1], m.viewing)
					m.FetchClaims(m.selections[m.viewing-1], m.viewing, 0, 20)
				} else {
					// Send claim for the selected IP
					// TODO: Implement claim sending logic
				}
			}

		case "tab":
			if m.inputSelected {
				m.inputSelected = false
				m.tableSelected = true
				m.nameInput.Blur()
				m.unitTables.Focus()
			} else {
				m.inputSelected = true
				m.tableSelected = false
				m.nameInput.Focus()
				m.unitTables.Blur()
			}
		}

	case string:
		if msg == "Claim sent successfully!" {
			m.statusMessage = statusMessageStyle.Render(msg)
			m.errorMessage = ""
		}

	case error:
		m.errorMessage = errorMessageStyle.Render(msg.Error())
		m.statusMessage = ""
	}

	// Update the selected row in the current table
	if m.tableSelected {
		t, cmd := m.unitTables[m.viewing].Update(msg)
		m.unitTables[m.viewing] = t
		cmds = append(cmds, cmd)
	} else if m.inputSelected {
		t, cmd := m.nameInput.Update(msg)
		m.nameInput = t
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if m.nameInput.Value() != "" {
			m.name = m.nameInput.Value()
		} else {
			m.name = "Anonymous"
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current state of the model
func (m *Model) View() string {
	activeTable := m.unitTables[m.viewing]
	m.FetchClaims(m.GetParentSelection(m.viewing), m.viewing, activeTable.Cursor()-activeTable.Height(), activeTable.Cursor()+activeTable.Height())

	msg := m.statusMessage
	if m.errorMessage != "" {
		msg = m.errorMessage
	}

	return titleStyle.Render("SpaceNet Browser") + "\n\n" +
		m.nameInput.View() + "\n\n" +
		tableStyle.Render(m.unitTables[m.viewing].View()) + "\n\n" +
		msg + "\n" +
		helpStyle("enter: select subnet, esc: back, q: quit")
}

func main() {
	// Parse command line flags
	server := flag.String("server", "::1", "IPv6 address of the server")
	httpPort := flag.Int("http-port", 8080, "HTTP port for the server's API")
	udpPort := flag.Int("port", 1337, "UDP port number of the server")
	flag.Parse()

	// Set up logging
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("Fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	// Initialize the TUI
	p := tea.NewProgram(Initialize(*server, *httpPort, *udpPort), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
