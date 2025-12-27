package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/go-toast/toast"
	"github.com/jchv/go-webview2"
)

// --- CONFIG ---
const (
	ConfigFile = "config.json"
	LogFile    = "app.log"
)

type Config struct {
	Ltuid  string `json:"ltuid_v2"`
	Ltoken string `json:"ltoken_v2"`
}

var (
	currentConfig Config
	menuStatus    *systray.MenuItem
)

// --- MAIN ---
func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// Icon Setup (Simple Box)
	systray.SetIcon(getIconBytes())
	systray.SetTitle("HoyoDaily Go")
	systray.SetTooltip("HoyoDaily Auto Check-in")

	// Menu
	menuStatus = systray.AddMenuItem("Status: Idle", "Current status")
	menuStatus.Disable()
	
	systray.AddSeparator()
	mCheck := systray.AddMenuItem("Check-in Now", "Force check-in")
	mLogin := systray.AddMenuItem("Login / Setup", "Open login window")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Exit", "Quit application")

	// Load Config
	loadConfig()
	
	// Start Scheduler (Goroutine)
	go scheduler()

	// Event Loop
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-mLogin.ClickedCh:
				go performLogin()
			case <-mCheck.ClickedCh:
				go performCheckIn()
			}
		}
	}()
}

func onExit() {
	// Clean exit
}

// --- LOGIC: SCHEDULER ---
func scheduler() {
	// Initial run after 5 seconds
	time.Sleep(5 * time.Second)
	performCheckIn()

	ticker := time.NewTicker(30 * time.Minute)
	for range ticker.C {
		performCheckIn()
	}
}

// --- LOGIC: CHECK-IN ---
func performCheckIn() {
	if currentConfig.Ltuid == "" {
		updateStatus("Status: Need Login")
		return
	}

	updateStatus("Status: Checking...")
	
	var results []string

	// Genshin
	gRes := claimReward("genshin", "e202102251931481", "https://sg-hk4e-api.hoyolab.com/event/sol/sign")
	results = append(results, "G: "+gRes)

	// Star Rail
	hRes := claimReward("starrail", "e202303301540311", "https://sg-public-api.hoyolab.com/event/luna/os/sign")
	results = append(results, "H: "+hRes)

	finalMsg := strings.Join(results, " | ")
	updateStatus("Status: Done")
	
	// Notify if success or error
	if strings.Contains(finalMsg, "✅") || strings.Contains(finalMsg, "❌") {
		pushNotification("HoyoDaily Report", finalMsg)
	}
}

func claimReward(gameAlias, actID, endpoint string) string {
	client := &http.Client{Timeout: 10 * time.Second}
	
	reqBody := strings.NewReader(fmt.Sprintf(`{"act_id": "%s", "lang": "en-us"}`, actID))
	req, _ := http.NewRequest("POST", endpoint, reqBody)
	
	// Headers
	req.Header.Set("Cookie", fmt.Sprintf("ltoken_v2=%s; ltuid_v2=%s;", currentConfig.Ltoken, currentConfig.Ltuid))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://act.hoyolab.com/")
	req.Header.Set("Origin", "https://act.hoyolab.com")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "❌ NetErr"
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	
	// Parse JSON Response
	var jsonRes map[string]interface{}
	json.Unmarshal(body, &jsonRes)

	retcode, ok := jsonRes["retcode"].(float64)
	if !ok {
		return "❌ APIErr"
	}

	if retcode == 0 {
		return "✅ Success"
	} else if retcode == -5003 {
		return "👌 Done" // Already claimed
	} else {
		return fmt.Sprintf("❌ Err(%.0f)", retcode)
	}
}

// --- LOGIC: LOGIN (WebView2) ---
func performLogin() {
	updateStatus("Status: Login Window Open...")

	w := webview2.New(true)
	defer w.Destroy()

	w.SetTitle("HoyoDaily Login - Silakan Login")
	w.SetSize(600, 700, webview2.HintNone)

	// Inject JS to grab cookies
	// We use a timer in JS to check document.cookie periodically
	jsScript := `
		setInterval(function() {
			var cookies = document.cookie;
			if (cookies.includes("ltuid_v2") && cookies.includes("ltoken_v2")) {
				window.chrome.webview.postMessage(cookies);
			}
		}, 1000);
	`
	
	w.Init(jsScript)
	w.Navigate("https://www.hoyolab.com/checkin-list")

	// Setup message handler from JS
	done := make(chan bool)
	
	w.Bind("window.chrome.webview.postMessage", func(cookies string) {
		// Parse cookies string
		ltuid := parseCookie(cookies, "ltuid_v2")
		ltoken := parseCookie(cookies, "ltoken_v2")
		
		if ltuid != "" && ltoken != "" {
			currentConfig.Ltuid = ltuid
			currentConfig.Ltoken = ltoken
			saveConfig()
			pushNotification("Login Sukses", "Cookie berhasil diambil!")
			updateStatus("Status: Login OK")
			w.Terminate() // Close window
			done <- true
		}
	})

	w.Run()
}

// --- UTILS ---

func parseCookie(cookieStr, key string) string {
	parts := strings.Split(cookieStr, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, key+"=") {
			return strings.TrimPrefix(p, key+"=")
		}
	}
	return ""
}

func loadConfig() {
	file, err := ioutil.ReadFile(ConfigFile)
	if err == nil {
		json.Unmarshal(file, &currentConfig)
	}
}

func saveConfig() {
	data, _ := json.MarshalIndent(currentConfig, "", "    ")
	ioutil.WriteFile(ConfigFile, data, 0644)
}

func updateStatus(msg string) {
	if menuStatus != nil {
		menuStatus.SetTitle(msg)
	}
}

func pushNotification(title, msg string) {
	notification := toast.Notification{
		AppID:   "HoyoDailyGo",
		Title:   title,
		Message: msg,
		Actions: []toast.Action{
			{"protocol", "Buka App", ""},
		},
	}
	notification.Push()
}

// Simple Icon (Blue Box)
func getIconBytes() []byte {
	// A simple blue pixel BMP header + data would be too long to hardcode properly here.
	// For simplicity, systray allows reading from file too, but we want embedded.
	// I will read a local "icon.ico" if exists, or assume blank (which works but is invisible).
	// Let's create a dummy file reader helper.
	
	// PROPER WAY: User needs an icon file. I will generate one in step 4.
	b, err := ioutil.ReadFile("icon.ico")
	if err != nil {
		return nil
	}
	return b
}
