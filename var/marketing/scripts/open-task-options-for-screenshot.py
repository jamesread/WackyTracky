#!/usr/bin/env python3
"""
Selenium script for use with repo-common-take-screenshot.py --script.

Opens the task details/options panel on the first task in the list by
right-clicking the first .task-row (context menu opens the task details overlay).
Call this after loading a list view URL.

Usage:
  python repo-common-take-screenshot.py task-options http://localhost:8080/lists/... \\
    --script var/marketing/scripts/open-task-options-for-screenshot.py
"""
import time

from selenium.webdriver.common.by import By
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC


def run(driver):
    wait = WebDriverWait(driver, timeout=10)
    first_row = wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, ".task-row")))
    actions = ActionChains(driver)
    actions.context_click(first_row).perform()
    time.sleep(0.5)
