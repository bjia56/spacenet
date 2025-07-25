// A TUI client for the SpaceNet server
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/bjia56/spacenet/api"
	"github.com/charmbracelet/bubbles/list"
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

// Import the shared API types from the api package

// Model represents the state of our application
type Model struct {
	serverAddr      string
	httpPort        int
	udpPort         int
	name            string
	sourceIP        string
	selectedSubnet  string
	prefixView      bool
	subnetView      bool
	claimView       bool
	prefixOptions   list.Model
	subnetTable     table.Model
	ipTable         table.Model
	textInput       textinput.Model
	statusMessage   string
	errorMessage    string
	currentPrefix   int
	currentBaseAddr string
	lastApiResponse *api.SubnetResponse
	prefixLengths   []int
}

// PrefixItem represents a prefix length item for the list
type PrefixItem struct {
	title       string
	description string
}

func (i PrefixItem) Title() string       { return i.title }
func (i PrefixItem) Description() string { return i.description }
func (i PrefixItem) FilterValue() string { return i.title }

// Initialize returns an initial model
func Initialize(serverAddr string, httpPort, udpPort int, name string, sourceIP string) Model {
	// Initialize text input for IP address entry
	ti := textinput.New()
	ti.Placeholder = "Enter IPv6 address to claim"
	ti.CharLimit = 45
	ti.Width = 45

	// Define prefix items
	prefixItems := []list.Item{
		PrefixItem{title: "/16", description: "Browse /16 subnets"},
		PrefixItem{title: "/32", description: "Browse /32 subnets"},
		PrefixItem{title: "/48", description: "Browse /48 subnets"},
		PrefixItem{title: "/64", description: "Browse /64 subnets"},
		PrefixItem{title: "/80", description: "Browse /80 subnets"},
		PrefixItem{title: "/96", description: "Browse /96 subnets"},
		PrefixItem{title: "/112", description: "Browse /112 subnets"},
		PrefixItem{title: "/128", description: "Browse /128 subnets"},
	}

	// Initialize list model
	prefixList := list.New(prefixItems, list.NewDefaultDelegate(), 0, 0)
	prefixList.Title = "Select Prefix Length"

	// Initialize table model for subnets
	columns := []table.Column{
		{Title: "Subnet", Width: 25},
		{Title: "Total", Width: 15},
		{Title: "Claimed", Width: 15},
		{Title: "Dominant Player", Width: 25},
		{Title: "Percentage", Width: 10},
	}

	subnetTable := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Initialize table model for IP addresses
	ipColumns := []table.Column{
		{Title: "IP", Width: 40},
		{Title: "Owner", Width: 40},
	}

	ipTable := table.New(
		table.WithColumns(ipColumns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	return Model{
		serverAddr:    serverAddr,
		httpPort:      httpPort,
		udpPort:       udpPort,
		name:          name,
		sourceIP:      sourceIP,
		prefixView:    true,
		subnetView:    false,
		claimView:     false,
		prefixOptions: prefixList,
		subnetTable:   subnetTable,
		ipTable:       ipTable,
		textInput:     ti,
		prefixLengths: []int{16, 32, 48, 64, 80, 96, 112, 128},
		statusMessage: "Welcome to SpaceNet! Select a prefix to browse.",
	}
}

// UpdateSubnetTable updates the subnet table with data from the API
func (m *Model) UpdateSubnetTable(prefixLen int, baseAddr string) tea.Cmd {
	return func() tea.Msg {
		var rows []table.Row

		// If base address is empty, we'll display root subnets
		// This is a placeholder - in a real implementation, you might get a list of top-level subnets
		// For now, we'll just show some example subnets
		if baseAddr == "" {
			// You would fetch this data from the server, but for now we'll just use examples
			for i := 1; i <= 5; i++ {
				subnet := fmt.Sprintf("2001:db8:%d::%d/%d", i, i, prefixLen)
				rows = append(rows, table.Row{
					subnet,
					"1000",
					"100",
					"Player" + fmt.Sprintf("%d", i),
					"10.0%",
				})
			}
			m.subnetTable.SetRows(rows)
			m.currentPrefix = prefixLen
			return nil
		}

		// Make API call to get subnet information
		url := fmt.Sprintf("http://%s:%d/api/subnet/%s/%d", m.serverAddr, m.httpPort, baseAddr, prefixLen)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch subnet data: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("API error: %s", resp.Status)
		}

		var subnetResp api.SubnetResponse
		if err := json.NewDecoder(resp.Body).Decode(&subnetResp); err != nil {
			return fmt.Errorf("failed to decode API response: %v", err)
		}

		// Store the response for later use
		m.lastApiResponse = &subnetResp

		// Add the main subnet info to the table
		rows = append(rows, table.Row{
			subnetResp.Subnet,
			subnetResp.TotalAddresses,
			subnetResp.ClaimedAddresses,
			subnetResp.DominantClaimant,
			fmt.Sprintf("%.2f%%", subnetResp.DominantPercentage),
		})

		// Add rows for each claimant's percentage
		for claimant, percentage := range subnetResp.AllClaimants {
			if claimant != subnetResp.DominantClaimant {
				rows = append(rows, table.Row{
					"",
					"",
					"",
					claimant,
					fmt.Sprintf("%.2f%%", percentage),
				})
			}
		}

		m.subnetTable.SetRows(rows)
		m.currentPrefix = prefixLen
		m.currentBaseAddr = baseAddr
		return nil
	}
}

// FetchIPDetails fetches IP ownership details
func (m *Model) FetchIPDetails(ip string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("http://%s:%d/api/ip/%s", m.serverAddr, m.httpPort, ip)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch IP data: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			// IP is unclaimed
			return nil
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("API error: %s - %s", resp.Status, body)
		}

		var claimResp api.ClaimResponse
		if err := json.NewDecoder(resp.Body).Decode(&claimResp); err != nil {
			return fmt.Errorf("failed to decode API response: %v", err)
		}

		// Update the IP table with the response
		var rows []table.Row
		rows = append(rows, table.Row{claimResp.IP, claimResp.Claimant})
		m.ipTable.SetRows(rows)
		return nil
	}
}

// SendClaim sends a claim for an IP
func (m *Model) SendClaim(ip string) tea.Cmd {
	return func() tea.Msg {
		// Build sendip command
		cmd := []string{"sendip", "-d", m.name, "-p", "ipv6"}
		if m.sourceIP != "" {
			cmd = append(cmd, "-6s", m.sourceIP)
		}
		cmd = append(cmd, "-p", "udp", "-ud", fmt.Sprintf("%d", m.udpPort), ip)

		// Execute sendip command
		err := exec.Command("sudo", "sh", "-c", strings.Join(cmd, " ")).Run()
		if err != nil {
			return fmt.Errorf("failed to execute sendip command: %v", err)
		}

		return "Claim sent successfully!"
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			if m.claimView {
				m.claimView = false
				m.subnetView = true
				return m, nil
			}
			if m.subnetView {
				m.subnetView = false
				m.prefixView = true
				return m, nil
			}

		case "enter":
			if m.prefixView {
				_, ok := m.prefixOptions.SelectedItem().(PrefixItem)
				if ok {
					m.prefixView = false
					m.subnetView = true
					prefixLen := m.prefixLengths[m.prefixOptions.Index()]
					return m, m.UpdateSubnetTable(prefixLen, "")
				}
			} else if m.subnetView {
				if len(m.subnetTable.Rows()) > 0 {
					m.subnetView = false
					m.claimView = true
					selectedRow := m.subnetTable.SelectedRow()
					if len(selectedRow) > 0 && selectedRow[0] != "" {
						m.selectedSubnet = selectedRow[0]
						m.textInput.SetValue(strings.Split(m.selectedSubnet, "/")[0])
						m.textInput.Focus()
					}
				}
			} else if m.claimView {
				ip := m.textInput.Value()
				if ip == "" {
					m.errorMessage = "Please enter a valid IPv6 address"
					return m, nil
				}
				m.statusMessage = "Sending claim..."
				return m, tea.Batch(
					m.SendClaim(ip),
					m.FetchIPDetails(ip),
				)
			}

		case "tab":
			if m.claimView {
				m.textInput.Focus()
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

	// Handle list update
	if m.prefixView {
		newListModel, cmd := m.prefixOptions.Update(msg)
		m.prefixOptions = newListModel
		cmds = append(cmds, cmd)
	}

	// Handle table update
	if m.subnetView {
		newTableModel, cmd := m.subnetTable.Update(msg)
		m.subnetTable = newTableModel
		cmds = append(cmds, cmd)
	}

	// Handle text input update
	if m.claimView {
		newTextInput, cmd := m.textInput.Update(msg)
		m.textInput = newTextInput
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current state of the model
func (m Model) View() string {
	if m.prefixView {
		return titleStyle.Render("SpaceNet Browser") + "\n\n" +
			m.prefixOptions.View() + "\n\n" +
			helpStyle("q: quit")
	}

	if m.subnetView {
		return titleStyle.Render("SpaceNet Browser - Subnet View") + "\n\n" +
			tableStyle.Render(m.subnetTable.View()) + "\n\n" +
			m.statusMessage + "\n" + m.errorMessage + "\n\n" +
			helpStyle("enter: select subnet, esc: back, q: quit")
	}

	if m.claimView {
		var view strings.Builder
		view.WriteString(titleStyle.Render("SpaceNet Browser - Claim IP") + "\n\n")
		view.WriteString("Selected Subnet: " + m.selectedSubnet + "\n\n")
		view.WriteString("Enter IPv6 address to claim: \n")
		view.WriteString(m.textInput.View() + "\n\n")

		// If we have IP details, show them
		if len(m.ipTable.Rows()) > 0 {
			view.WriteString("Current owner: \n")
			view.WriteString(tableStyle.Render(m.ipTable.View()) + "\n\n")
		}

		view.WriteString(m.statusMessage + "\n" + m.errorMessage + "\n\n")
		view.WriteString(helpStyle("enter: send claim, esc: back, tab: focus input, q: quit"))
		return view.String()
	}

	return "Loading..."
}

func main() {
	// Parse command line flags
	name := flag.String("name", "Anonymous", "Your name for the claim")
	server := flag.String("server", "::1", "IPv6 address of the server")
	httpPort := flag.Int("http-port", 8080, "HTTP port for the server's API")
	udpPort := flag.Int("port", 1337, "UDP port number of the server")
	source := flag.String("source", "", "Source IP address to claim (optional)")
	flag.Parse()

	// Set up logging
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("Fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	// Initialize the TUI
	p := tea.NewProgram(Initialize(*server, *httpPort, *udpPort, *name, *source), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
