"""Opens the task details dialog for the first task (right-click opens it)."""
import time

from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait


def run(driver):
    row = WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.CSS_SELECTOR, ".task-row")))
    ActionChains(driver).context_click(row).perform()
    WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.CSS_SELECTOR, ".task-details-dialog")))
    time.sleep(0.75)
