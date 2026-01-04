// Package app implements the main application model and view navigation logic.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	"github.com/0xjuanma/golazo/internal/notify"
	"github.com/0xjuanma/golazo/internal/reddit"
	"github.com/0xjuanma/golazo/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// view represents the current application view.
type view int

const (
	viewMain view = iota
	viewLiveMatches
	viewStats
	viewSettings
)

// model holds the application state.
// Fields are organized by concern: display, data, UI components, and configuration.
type model struct {
	// Display dimensions
	width  int
	height int

	// View state
	currentView view
	selected    int

	// Match data
	matches             []ui.MatchDisplay
	upcomingMatches     []ui.MatchDisplay // Upcoming matches for 1-day stats view (deprecated, kept for compatibility)
	liveUpcomingMatches []ui.MatchDisplay // Upcoming matches for live view (shown at bottom of left panel)
	matchDetails        *api.MatchDetails
	matchDetailsCache   map[int]*api.MatchDetails // Cache to avoid repeated API calls
	liveUpdates         []string
	lastEvents          []api.MatchEvent
	lastHomeScore       int // Track last known home score for goal notifications
	lastAwayScore       int // Track last known away score for goal notifications

	// Stats data cache - stores 5 days of data, filtered client-side for Today/3d/5d views
	statsData *fotmob.StatsData

	// Progressive loading state (stats view)
	statsDaysLoaded int // Number of days loaded so far (0-5)
	statsTotalDays  int // Total days to load (5)

	// Progressive loading state (live view) - batch-based for parallel fetching
	liveBatchesLoaded int         // Number of batches loaded so far
	liveTotalBatches  int         // Total batches to load
	liveMatchesBuffer []api.Match // Buffer to accumulate live matches during progressive load

	// UI components
	spinner          spinner.Model
	randomSpinner    *ui.RandomCharSpinner
	statsViewSpinner *ui.RandomCharSpinner // Separate spinner for stats view
	pollingSpinner   *ui.RandomCharSpinner // Small spinner for polling indicator

	// List components
	liveMatchesList     list.Model
	statsMatchesList    list.Model
	upcomingMatchesList list.Model

	// Loading states
	loading          bool
	mainViewLoading  bool
	liveViewLoading  bool
	statsViewLoading bool
	polling          bool
	pendingSelection int // Tracks which view is being preloaded (-1 = none, 0 = stats, 1 = live)

	// Configuration
	useMockData    bool
	debugMode      bool // Enable debug logging to file
	statsDateRange int  // 1, 3, or 5 days (default: 1)

	// Settings view state
	settingsState *ui.SettingsState

	// API clients
	fotmobClient *fotmob.Client
	parser       *fotmob.LiveUpdateParser
	redditClient *reddit.Client

	// Goal replay links from Reddit (keyed by matchID:minute)
	goalLinks map[reddit.GoalLinkKey]*reddit.GoalLink

	// Pending goals to fetch links for (keyed by matchID:minute)
	// Stores full GoalInfo for efficient background processing
	pendingGoals map[reddit.GoalLinkKey]reddit.GoalInfo

	// Notifications
	notifier *notify.DesktopNotifier
}

// New creates a new application model with default values.
// useMockData determines whether to use mock data instead of real API data.
// debugMode enables debug logging to a file.
func New(useMockData bool, debugMode bool) model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = ui.SpinnerStyle()

	// Initialize random character spinners
	randomSpinner := ui.NewRandomCharSpinner()
	randomSpinner.SetWidth(30)

	statsViewSpinner := ui.NewRandomCharSpinner()
	statsViewSpinner.SetWidth(30)

	pollingSpinner := ui.NewRandomCharSpinner()
	pollingSpinner.SetWidth(10) // Small spinner for polling indicator

	// Initialize list models with custom delegate
	delegate := ui.NewMatchListDelegate()

	// Filter input styles matching neon theme
	filterCursorStyle, filterPromptStyle := ui.FilterInputStyles()

	liveList := list.New([]list.Item{}, delegate, 0, 0)
	liveList.SetShowTitle(false)
	liveList.SetShowStatusBar(true)
	liveList.SetFilteringEnabled(true)
	liveList.SetShowFilter(true)
	liveList.Filter = list.DefaultFilter // Required for filtering to work
	liveList.Styles.FilterCursor = filterCursorStyle
	liveList.FilterInput.PromptStyle = filterPromptStyle
	liveList.FilterInput.Cursor.Style = filterCursorStyle

	statsList := list.New([]list.Item{}, delegate, 0, 0)
	statsList.SetShowTitle(false)
	statsList.SetShowStatusBar(true)
	statsList.SetFilteringEnabled(true)
	statsList.SetShowFilter(true)
	statsList.Filter = list.DefaultFilter // Required for filtering to work
	statsList.Styles.FilterCursor = filterCursorStyle
	statsList.FilterInput.PromptStyle = filterPromptStyle
	statsList.FilterInput.Cursor.Style = filterCursorStyle

	upcomingList := list.New([]list.Item{}, delegate, 0, 0)
	upcomingList.SetShowTitle(false)
	upcomingList.SetShowStatusBar(true)
	upcomingList.SetFilteringEnabled(true)
	upcomingList.SetShowFilter(true)
	upcomingList.Filter = list.DefaultFilter // Required for filtering to work
	upcomingList.Styles.FilterCursor = filterCursorStyle
	upcomingList.FilterInput.PromptStyle = filterPromptStyle
	upcomingList.FilterInput.Cursor.Style = filterCursorStyle

	// Initialize Reddit client (best-effort, nil if fails)
	var redditClient *reddit.Client
	if debugMode {
		redditClient, _ = reddit.NewClientWithDebug(func(message string) {
			// This will be called by the Reddit client for debug logging
			// We'll create a model instance to access debugLog, but for now just log directly
			// This is a bit of a hack, but it works for debug logging
			configDir, _ := data.ConfigDir()
			if configDir != "" {
				logFile := filepath.Join(configDir, "golazo_debug.log")
				f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					defer f.Close()
					f.WriteString(fmt.Sprintf("%s %s\n", time.Now().Format("2006-01-02 15:04:05"), message))
				}
			}
		})
	} else {
		redditClient, _ = reddit.NewClient()
	}

	return model{
		currentView:         viewMain,
		matchDetailsCache:   make(map[int]*api.MatchDetails),
		useMockData:         useMockData,
		debugMode:           debugMode,
		fotmobClient:        fotmob.NewClient(),
		parser:              fotmob.NewLiveUpdateParser(),
		redditClient:        redditClient,
		goalLinks:           make(map[reddit.GoalLinkKey]*reddit.GoalLink),
		pendingGoals:        make(map[reddit.GoalLinkKey]reddit.GoalInfo),
		notifier:            notify.NewDesktopNotifier(),
		spinner:             s,
		randomSpinner:       randomSpinner,
		statsViewSpinner:    statsViewSpinner,
		pollingSpinner:      pollingSpinner,
		liveMatchesList:     liveList,
		statsMatchesList:    statsList,
		upcomingMatchesList: upcomingList,
		statsDateRange:      1,
		pendingSelection:    -1, // No pending selection
	}
}

// Init initializes the application.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, ui.SpinnerTick())
}
