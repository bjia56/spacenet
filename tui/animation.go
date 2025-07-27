package main

import (
	"crypto/md5"
	"encoding/binary"
	"io"
	"math/rand"
	"net"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type Animation interface {
	tea.Model
	Tick()
	SetDimensions(width, height int)
	AnimateForIP(ip net.IP)
	RandBytes(n int) []byte
}

// DefaultAnimation is to be embedded in other animations to define
// default behavior for SetDimensions and select tea.Model methods.
type DefaultAnimation struct {
	width  int
	height int

	parent Animation
	ip     net.IP
	rand   *rand.Rand
}

func NewDefaultAnimation(parent Animation) *DefaultAnimation {
	d := &DefaultAnimation{
		width:  80,
		height: 24,
		parent: parent,
	}
	d.AnimateForIP(net.IPv6zero)
	return d
}

func (d *DefaultAnimation) SetDimensions(width, height int) {
	d.width = width
	d.height = height
	if d.width < 10 {
		d.width = 10 // Minimum width
	}
	if d.height < 10 {
		d.height = 10 // Minimum height
	}
}

func (d *DefaultAnimation) Init() tea.Cmd {
	return nil
}

func (d *DefaultAnimation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case timer.TickMsg:
		d.parent.Tick()
		return d.parent, nil
	}
	return d.parent, nil
}

func (d *DefaultAnimation) AnimateForIP(ip net.IP) {
	if ip.Equal(d.ip) {
		return // No change
	}
	d.ip = ip
	h := md5.New()
	io.WriteString(h, ip.String())
	d.rand = rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(h.Sum(nil)))))
}

func (d *DefaultAnimation) RandBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(d.rand.Intn(256))
	}
	return b
}
