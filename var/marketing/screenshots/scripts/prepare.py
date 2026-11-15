"""Waits for the SPA to render data before capture."""
import time

from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys


def run(driver):
    deadline = time.time() + 15
    while time.time() < deadline:
        if driver.find_elements(By.CSS_SELECTOR, "main .task-row, main a[href^='/lists/'], main form, main input"):
            break
        time.sleep(0.25)
    driver.execute_script("document.activeElement && document.activeElement.blur(); window.scrollTo(0, 0);")
    # Lists focus their first task on load; Escape clears it so the add bar shows its normal prompt.
    ActionChains(driver).send_keys(Keys.ESCAPE).perform()
    time.sleep(0.3)
