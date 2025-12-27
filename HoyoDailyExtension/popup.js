document.addEventListener('DOMContentLoaded', () => {
  // Load saved status from storage
  chrome.storage.local.get(['status', 'lastCheck'], (result) => {
    document.getElementById('status').textContent = result.status || "No data available";
    document.getElementById('time').textContent = result.lastCheck || "Never";
  });

  // Manual Check Button Logic
  document.getElementById('btnCheck').addEventListener('click', () => {
    const btn = document.getElementById('btnCheck');
    const msg = document.getElementById('msg');
    
    // UI Feedback: Loading state
    btn.disabled = true;
    btn.textContent = "Checking...";
    msg.textContent = "";
    
    // Send message to background script
    chrome.runtime.sendMessage({ action: "manualCheck" }, (response) => {
      // Small delay to allow storage to update
      setTimeout(() => {
        chrome.storage.local.get(['status', 'lastCheck'], (result) => {
          // Update UI with new data
          document.getElementById('status').textContent = result.status;
          document.getElementById('time').textContent = result.lastCheck;
          
          // Reset button state
          btn.disabled = false;
          btn.textContent = "Check In Now";
          
          // Success feedback
          msg.textContent = "Check-in Complete!";
          setTimeout(() => { msg.textContent = ""; }, 3000);
        });
      }, 1000);
    });
  });
});