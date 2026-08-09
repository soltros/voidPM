package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/voidlinux/voidpm/pkg/kernel"
	"github.com/voidlinux/voidpm/pkg/runit"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

type Tab int

const (
	TabServices Tab = iota
	TabPackages
	TabSrc
	TabMaintenance
)

// ServiceItem implements list.Item for runit services
type ServiceItem struct {
	srv *runit.Service
}

func (i ServiceItem) Title() string {
	statusBadge := ""
	switch i.srv.Status {
	case runit.StatusRunning:
		statusBadge = ui.RunningBadge.Render(" RUNNING ")
	case runit.StatusStopped:
		statusBadge = ui.StoppedBadge.Render(" STOPPED ")
	case runit.StatusDisabled:
		statusBadge = ui.DisabledBadge.Render(" DISABLED ")
	default:
		statusBadge = ui.ErrorBadge.Render(" " + string(i.srv.Status) + " ")
	}
	return fmt.Sprintf("%-25s %s", i.srv.Name, statusBadge)
}

func (i ServiceItem) Description() string {
	if i.srv.Status == runit.StatusRunning {
		return fmt.Sprintf("PID: %d | Uptime: %s | Path: %s", i.srv.PID, i.srv.FormattedUptime(), i.srv.ActiveLink)
	}
	if i.srv.Enabled {
		return fmt.Sprintf("Enabled in %s (Stopped)", i.srv.ActiveLink)
	}
	return fmt.Sprintf("Available in %s", i.srv.ServiceDir)
}

func (i ServiceItem) FilterValue() string {
	return i.srv.Name
}

// PackageItem implements list.Item for XBPS packages
type PackageItem struct {
	pkg xbps.Package
}

func (i PackageItem) Title() string {
	status := "[   ]"
	if i.pkg.Installed {
		status = ui.InstalledBadge.Render("[INST]")
	}
	if i.pkg.Orphan {
		status += " " + ui.OrphanBadge.Render("[ORPHAN]")
	}
	return fmt.Sprintf("%s %-20s %s", status, i.pkg.Name, i.pkg.Version)
}

func (i PackageItem) Description() string {
	return i.pkg.ShortDesc
}

func (i PackageItem) FilterValue() string {
	return i.pkg.Name + " " + i.pkg.ShortDesc
}

type Model struct {
	activeTab     Tab
	svManager     *runit.Manager
	xbClient      *xbps.Client
	srcManager    *xbps.SrcManager
	cleaner       *xbps.Cleaner
	
	serviceList   list.Model
	packageList   list.Model
	searchInput   textinput.Model
	logViewport   viewport.Model
	
	width         int
	height        int
	statusMessage string
	showLogs      bool
	currentLogSrv string
}

func NewModel() (*Model, error) {
	svMgr := runit.NewManager()
	xbCl := xbps.NewClient()
	srcMgr := xbps.NewSrcManager()
	cleaner := xbps.NewCleaner()

	// Initialize service list
	srvs, _ := svMgr.ListServices()
	items := make([]list.Item, len(srvs))
	for i, s := range srvs {
		items[i] = ServiceItem{srv: s}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Runit Services"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	// Initialize package search input & list
	ti := textinput.New()
	ti.Placeholder = "Search xbps packages (press Enter)..."
	ti.CharLimit = 156
	ti.Width = 40

	pkgList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	pkgList.Title = "XBPS Packages"
	pkgList.SetShowStatusBar(true)

	vp := viewport.New(80, 20)

	m := &Model{
		activeTab:     TabServices,
		svManager:     svMgr,
		xbClient:      xbCl,
		srcManager:    srcMgr,
		cleaner:       cleaner,
		serviceList:   l,
		packageList:   pkgList,
		searchInput:   ti,
		logViewport:   vp,
		statusMessage: "Press Tab to switch views | Space: start/stop | e: enable/disable | r: restart | l: logs | q: quit",
	}

	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := msg.Height - 8
		if h < 5 {
			h = 5
		}
		m.serviceList.SetSize(msg.Width-4, h)
		m.packageList.SetSize(msg.Width-4, h)
		m.logViewport.Width = msg.Width - 4
		m.logViewport.Height = h

	case tea.KeyMsg:
		// Global keys when search input is not focused
		if !m.searchInput.Focused() && !m.serviceList.SettingFilter() {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "tab":
				m.activeTab = (m.activeTab + 1) % 4
				if m.activeTab == TabPackages {
					m.reloadInstalledPackages()
				}
				return m, nil
			case "1":
				m.activeTab = TabServices
				return m, nil
			case "2":
				m.activeTab = TabPackages
				m.reloadInstalledPackages()
				return m, nil
			case "3":
				m.activeTab = TabSrc
				return m, nil
			case "4":
				m.activeTab = TabMaintenance
				return m, nil
			}
		}

		// Tab-specific keybindings
		switch m.activeTab {
		case TabServices:
			if m.showLogs {
				if msg.String() == "esc" || msg.String() == "l" || msg.String() == "q" {
					m.showLogs = false
					return m, nil
				}
				var vpCmd tea.Cmd
				m.logViewport, vpCmd = m.logViewport.Update(msg)
				return m, vpCmd
			}

			if !m.serviceList.SettingFilter() {
				sel, ok := m.serviceList.SelectedItem().(ServiceItem)
				if ok && sel.srv != nil {
					switch msg.String() {
					case " ":
						if sel.srv.Status == runit.StatusRunning {
							err := m.svManager.Stop(sel.srv.Name)
							m.setStatusErr("Stop", sel.srv.Name, err)
						} else {
							err := m.svManager.Start(sel.srv.Name)
							m.setStatusErr("Start", sel.srv.Name, err)
						}
						m.reloadServices()
						return m, nil
					case "e":
						if sel.srv.Enabled {
							err := m.svManager.DisableService(sel.srv.Name)
							m.setStatusErr("Disable", sel.srv.Name, err)
						} else {
							err := m.svManager.EnableService(sel.srv.Name)
							m.setStatusErr("Enable", sel.srv.Name, err)
						}
						m.reloadServices()
						return m, nil
					case "r":
						err := m.svManager.Restart(sel.srv.Name)
						m.setStatusErr("Restart", sel.srv.Name, err)
						m.reloadServices()
						return m, nil
					case "l":
						lines, err := m.svManager.TailLogs(sel.srv.Name, 100)
						if err != nil {
							m.statusMessage = ui.RenderError("Log error: " + err.Error())
						} else {
							m.showLogs = true
							m.currentLogSrv = sel.srv.Name
							m.logViewport.SetContent(strings.Join(lines, "\n"))
							m.logViewport.GotoBottom()
						}
						return m, nil
					}
				}
			}

			var sCmd tea.Cmd
			m.serviceList, sCmd = m.serviceList.Update(msg)
			cmds = append(cmds, sCmd)

		case TabPackages:
			if m.searchInput.Focused() {
				if msg.String() == "enter" {
					m.searchInput.Blur()
					q := m.searchInput.Value()
					if q != "" {
						pkgs, err := m.xbClient.Search(q)
						if err != nil {
							m.statusMessage = ui.RenderError("Search error: " + err.Error())
						} else {
							items := make([]list.Item, len(pkgs))
							for i, p := range pkgs {
								items[i] = PackageItem{pkg: p}
							}
							m.packageList.SetItems(items)
							m.statusMessage = fmt.Sprintf("Found %d packages for '%s'", len(pkgs), q)
						}
					}
					return m, nil
				}
				var tiCmd tea.Cmd
				m.searchInput, tiCmd = m.searchInput.Update(msg)
				return m, tiCmd
			}

			if msg.String() == "/" {
				m.searchInput.Focus()
				return m, nil
			}

			var pCmd tea.Cmd
			m.packageList, pCmd = m.packageList.Update(msg)
			cmds = append(cmds, pCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) setStatusErr(action, name string, err error) {
	if err != nil {
		m.statusMessage = ui.RenderError(fmt.Sprintf("%s %s failed: %v", action, name, err))
	} else {
		m.statusMessage = ui.RenderSuccess(fmt.Sprintf("%s %s succeeded", action, name))
	}
}

func (m *Model) reloadServices() {
	srvs, err := m.svManager.ListServices()
	if err == nil {
		items := make([]list.Item, len(srvs))
		for i, s := range srvs {
			items[i] = ServiceItem{srv: s}
		}
		m.serviceList.SetItems(items)
	}
}

func (m *Model) reloadInstalledPackages() {
	pkgs, err := m.xbClient.ListInstalled()
	if err == nil {
		items := make([]list.Item, len(pkgs))
		for i, p := range pkgs {
			items[i] = PackageItem{pkg: p}
		}
		m.packageList.SetItems(items)
	}
}

func (m *Model) View() string {
	var sb strings.Builder

	// Header tabs
	tabs := []string{" [1] Services ", " [2] Packages ", " [3] Void-Src ", " [4] Maintenance "}
	for i, t := range tabs {
		if Tab(i) == m.activeTab {
			sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ui.BgDark).Background(ui.AccentColor).Render(t) + " ")
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(ui.FgLight).Background(ui.BgDark).Render(t) + " ")
		}
	}
	sb.WriteString("\n\n")

	// View content
	switch m.activeTab {
	case TabServices:
		if m.showLogs {
			sb.WriteString(ui.TitleStyle.Render("Logs for service: "+m.currentLogSrv) + " (Press Esc/q to exit)\n")
			sb.WriteString(m.logViewport.View())
		} else {
			sb.WriteString(m.serviceList.View())
		}
	case TabPackages:
		sb.WriteString("Search: " + m.searchInput.View() + "\n\n")
		sb.WriteString(m.packageList.View())
	case TabSrc:
		sb.WriteString(ui.TitleStyle.Render("Void Source Packages (xbps-src)") + "\n\n")
		if m.srcManager.IsSetup() {
			sb.WriteString(ui.RenderSuccess("Repository initialized at: "+m.srcManager.RepoDir) + "\n")
			sb.WriteString("Use CLI subcommands:\n")
			sb.WriteString("  - vpm src sync        (git pull void-packages)\n")
			sb.WriteString("  - vpm src build <pkg> (build template from source)\n")
			sb.WriteString("  - vpm src install <pkg> (install compiled .xbps package)\n")
		} else {
			sb.WriteString(ui.RenderWarning("void-packages repo is not set up.") + "\n")
			sb.WriteString("Run 'vpm src setup' or press enter in CLI to clone void-packages.\n")
		}
	case TabMaintenance:
		sb.WriteString(ui.TitleStyle.Render("Kernel Management & System Maintenance") + "\n\n")
		
		sk, err := kernel.GetSystemKernels()
		if err == nil {
			sb.WriteString(fmt.Sprintf("Running Kernel: %s\n", ui.InstalledBadge.Render(sk.RunningKernel)))
			if len(sk.OldPurgeable) > 0 {
				sb.WriteString(ui.RenderWarning(fmt.Sprintf("Purgeable Old Kernels: %v", sk.OldPurgeable)) + "\n")
			} else {
				sb.WriteString(ui.RenderSuccess("No obsolete kernels to purge") + "\n")
			}
			sb.WriteString("\nInstalled Kernel Packages:\n")
			for _, k := range sk.Installed {
				runText := ""
				if k.IsRunning {
					runText = " " + ui.RunningBadge.Render("RUNNING")
				}
				sb.WriteString(fmt.Sprintf("  - %-15s %s%s\n", k.Name, k.Version, runText))
			}
		}

		sb.WriteString("\nCLI Kernel Commands:\n")
		sb.WriteString("  - vpm kernel status              (View running/installed kernels)\n")
		sb.WriteString("  - vpm kernel switch linux-lts    (Switch to LTS kernel branch)\n")
		sb.WriteString("  - vpm kernel reconfigure         (Reconfigure initramfs/bootloader)\n")
		sb.WriteString("  - vpm kernel dracut              (Regenerate all initramfs images)\n")
		sb.WriteString("  - vpm kernel purge               (Remove obsolete kernels via vkpurge)\n")
		sb.WriteString("  - vpm kernel hold / unhold       (Pin kernel major version)\n")
		sb.WriteString("  - vpm clean all                  (Full system & cache cleanup)\n")
	}

	// Status line
	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(ui.MutedColor).Render("────────────────────────────────────────────────────────────────────────") + "\n")
	sb.WriteString(m.statusMessage + "\n")

	_ = key.NewBinding() // keeps key import clean if needed
	return sb.String()
}
