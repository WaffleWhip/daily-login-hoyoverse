// --- CONFIGURATION ---
const API_ENDPOINTS = {
  genshin: {
    url: 'https://sg-hk4e-api.hoyolab.com/event/sol/sign?lang=en-us',
    act_id: 'e202102251931481',
    name: 'Genshin'
  },
  starrail: {
    url: 'https://sg-public-api.hoyolab.com/event/luna/os/sign?lang=en-us',
    act_id: 'e202303301540311',
    name: 'Star Rail'
  }
};

// --- ALARM SCHEDULER ---
// Trigger alarm on install or browser startup
chrome.runtime.onInstalled.addListener(() => {
  console.log("HoyoDaily Installed. Initializing scheduler...");
  chrome.alarms.create('dailyCheck', { periodInMinutes: 60 }); // Check every hour
  performCheckIn(); // Initial check
});

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === 'dailyCheck') {
    performCheckIn();
  }
});

// --- CORE FUNCTION ---
async function performCheckIn() {
  console.log("Starting Check-in Process...");
  
  // 1. Retrieve Cookies
  const cookies = await getHoyoCookies();
  if (!cookies) {
    console.warn("No cookies found. User might not be logged in.");
    chrome.storage.local.set({ status: "❌ Failed: Not logged in at Hoyolab.com" });
    return;
  }

  // 2. Perform Requests
  const results = [];
  
  // Genshin Impact
  const gRes = await claimReward(cookies, API_ENDPOINTS.genshin);
  results.push(`Genshin Impact: ${gRes}`);

  // Honkai: Star Rail
  const hRes = await claimReward(cookies, API_ENDPOINTS.starrail);
  results.push(`Honkai: Star Rail: ${hRes}`);

  // 3. Update Status & Notify
  const finalStatus = results.join(" | ");
  const timestamp = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  
  chrome.storage.local.set({ 
    status: finalStatus,
    lastCheck: timestamp
  });

  // Notification Logic: Only notify on Success (New Claim) or Error. 
  // Stay silent if "Already Claimed" (just "✅") to avoid spamming the user every hour.
  // We check for "✅ Success" explicitly to distinguish from the simple "✅".
  if (finalStatus.includes("✅ Success") || finalStatus.includes("❌")) {
    chrome.notifications.create({
      type: 'basic',
      iconUrl: 'icon.png',
      title: 'HoyoDaily Report',
      message: finalStatus,
      priority: 2
    });
  }
}

async function getHoyoCookies() {
  try {
    const ltuid = await chrome.cookies.get({ url: "https://www.hoyolab.com", name: "ltuid_v2" });
    const ltoken = await chrome.cookies.get({ url: "https://www.hoyolab.com", name: "ltoken_v2" });

    if (ltuid && ltoken) {
      return { ltuid: ltuid.value, ltoken: ltoken.value };
    }
    
    // Fallback to V1 cookies (Legacy support)
    const ltuid_v1 = await chrome.cookies.get({ url: "https://www.hoyolab.com", name: "ltuid" });
    const ltoken_v1 = await chrome.cookies.get({ url: "https://www.hoyolab.com", name: "ltoken" });
    
    if (ltuid_v1 && ltoken_v1) {
      return { ltuid: ltuid_v1.value, ltoken: ltoken_v1.value };
    }
  } catch (e) {
    console.error("Error retrieving cookies:", e);
  }
  return null;
}

async function claimReward(cookies, gameConfig) {
  try {
    // Construct valid cookie string for the API
    const cookieStr = `ltoken_v2=${cookies.ltoken}; ltuid_v2=${cookies.ltuid};`;
    
    const response = await fetch(gameConfig.url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Referer': 'https://act.hoyolab.com/',
        'Origin': 'https://act.hoyolab.com',
        'User-Agent': navigator.userAgent
      },
      // Note: 'host_permissions' in manifest allows us to bypass some CORS restrictions
      body: JSON.stringify({ act_id: gameConfig.act_id })
    });

    const data = await response.json();
    
    if (data.retcode === 0) return "✅ Success";
    if (data.retcode === -5003) return "✅"; // Already claimed
    return `❌ Err(${data.retcode})`;
    
  } catch (error) {
    console.error(`Network Error (${gameConfig.name}):`, error);
    return "❌ NetErr";
  }
}

// Listener for messages from Popup (Manual Check Button)
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "manualCheck") {
    performCheckIn().then(() => {
      sendResponse({ status: "Done" });
    });
    return true; // Indicates asynchronous response
  }
});
