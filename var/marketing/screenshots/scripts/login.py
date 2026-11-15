"""Signs in with the throwaway stack's default admin/admin account, once.

The IAM database only exists under tmp/ for the duration of `make`. The
browser session is shared across captures, so later sections are already
signed in and this returns immediately.
"""
import time

from selenium.webdriver.common.by import By


def run(driver):
    deadline = time.time() + 10
    while time.time() < deadline:
        if driver.find_elements(By.ID, "username"):
            break
        if driver.find_elements(By.CSS_SELECTOR, "header .user-info"):
            return
        time.sleep(0.25)
    else:
        return
    driver.find_element(By.ID, "username").send_keys("admin")
    driver.find_element(By.ID, "password").send_keys("admin")
    driver.find_element(By.CSS_SELECTOR, "button[type=submit]").click()
    # Reload the requested route now that the session cookie is set.
    time.sleep(1.5)
    driver.refresh()
    time.sleep(1.5)
