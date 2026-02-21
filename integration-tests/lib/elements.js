import { By } from 'selenium-webdriver'
import fs from 'fs'
import { Condition } from 'selenium-webdriver'

export function takeScreenshotOnFailure (test, driver) {
  if (test.state === 'failed') {
    const title = test.fullTitle().replace(/[^\w\s-]/g, '_').replace(/\s+/g, '_')
    console.log('Test failed, taking screenshot: ' + title)
    driver.takeScreenshot().then((img) => {
      fs.mkdirSync('screenshots', { recursive: true })
      fs.writeFileSync('screenshots/' + title + '.failed.png', img, 'base64')
    })
  }
}

export async function getRootAndWait () {
  await webdriver.get(runner.baseUrl())
  await webdriver.wait(
    new Condition('page has title', async () => {
      const title = await webdriver.getTitle()
      return title != null && title.length > 0
    }),
    10000
  )
}
