import threading
import pyautogui
import time
import random
from pynput import keyboard

def main():
    press_any_key = threading.Event()
    threading.Thread(target=move_mouse, args=(press_any_key,), daemon=True).start()
    def on_press(_key) -> bool:
        press_any_key.set()
        return False
    try:
        with keyboard.Listener(on_press=on_press) as listener:
            listener.join()
            print("Key pressed. Exiting program.")
    except KeyboardInterrupt:
        print("Program terminated by user.")
    finally:
        press_any_key.set()


def move_mouse(stop_event):
    while not stop_event.is_set():
        screen = pyautogui.size()
        x = random.randint(0, screen.width - 1)
        y = random.randint(0, screen.height - 1)
        pyautogui.moveTo(x, y, duration=0.5)
        time.sleep(random.uniform(1, 5))

if __name__ == "__main__":
    main()
