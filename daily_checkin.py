import asyncio
import genshin
import json
import os
import logging
from windows_toasts import WindowsToaster, ToastText1

# Config
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
CONFIG_FILE = os.path.join(BASE_DIR, "config.json")
LOG_FILE = os.path.join(BASE_DIR, "activity.log")

# Log Setup
logging.basicConfig(
    filename=LOG_FILE,
    level=logging.INFO,
    format='[%(asctime)s] %(levelname)s: %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)

def notify(title, body):
    try:
        toaster = WindowsToaster('HoyoDaily')
        new_toast = ToastText1()
        new_toast.SetBody(f"{title}\n{body}")
        toaster.show_toast(new_toast)
    except:
        pass

async def main():
    if not os.path.exists(CONFIG_FILE):
        return

    try:
        with open(CONFIG_FILE, "r") as f:
            cookies = json.load(f)
    except:
        return

    client = genshin.Client(cookies)
    
    # Check Genshin
    g_status = "Skipped"
    try:
        reward = await client.claim_daily_reward(game=genshin.Game.GENSHIN)
        g_status = f"✅ +{reward.amount} {reward.name}"
    except genshin.AlreadyClaimed:
        g_status = "👌 Done"
    except Exception:
        g_status = "❌ Error"

    # Check Star Rail
    h_status = "Skipped"
    try:
        reward = await client.claim_daily_reward(game=genshin.Game.STARRAIL)
        h_status = f"✅ +{reward.amount} {reward.name}"
    except genshin.AlreadyClaimed:
        h_status = "👌 Done"
    except Exception:
        h_status = "❌ Error"

    # Notify only if something interesting happened
    log_msg = f"Genshin: {g_status} | HSR: {h_status}"
    logging.info(log_msg)
    
    if "✅" in log_msg or "❌" in log_msg:
        notify("HoyoDaily Report", log_msg)

if __name__ == "__main__":
    asyncio.run(main())
