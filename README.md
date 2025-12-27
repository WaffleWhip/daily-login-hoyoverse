# HoyoDaily Auto Check-in (Chrome Extension)

A lightweight, automated Chrome Extension to claim daily rewards for **Genshin Impact** and **Honkai: Star Rail** via Hoyolab.

## Features
*   **Zero Configuration**: Uses your browser's existing login session.
*   **Automated**: Runs in the background and checks for rewards every hour.
*   **Silent**: Only notifies you if a *new* reward is claimed or if an error occurs. (Does not spam "Already claimed" notifications).
*   **Manual Trigger**: "Check In Now" button available in the popup.
*   **Secure**: Runs entirely locally within your browser. No passwords are stored.

## Installation

Since this extension is not yet on the Chrome Web Store, you must install it in **Developer Mode**.

1.  **Clone or Download** this repository to a folder on your computer.
2.  Open **Google Chrome** (or Edge/Brave/Opera).
3.  Navigate to `chrome://extensions/` in the address bar.
4.  Toggle **Developer mode** on (usually in the top right corner).
5.  Click **Load unpacked**.
6.  Select the `HoyoDailyExtension` folder inside this repository.

## Usage

1.  Ensure you are logged in to [Hoyolab.com](https://www.hoyolab.com/) in your browser.
2.  The extension icon (purple box) will appear in your toolbar.
3.  Click the icon to see the last status.
4.  Click **Check In Now** to test it immediately.
5.  The extension will now automatically run in the background as long as your browser is open.

## Status Indicators
*   **✅ Success**: Successfully claimed a reward.
*   **👌 Done**: Reward was already claimed for today.
*   **❌ Failed**: Something went wrong (usually means you are logged out).
*   **❌ NetErr**: Network connection issue.

## Disclaimer
This project is not affiliated with Hoyoverse. Use at your own risk.
