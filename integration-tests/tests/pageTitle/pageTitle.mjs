import { describe, it, before, after, afterEach } from 'mocha'
import { expect } from 'chai'
import { getRootAndWait, takeScreenshotOnFailure } from '../../lib/elements.js'

describe('page title', function () {
  before(async function () {
    await runner.start()
  })

  after(async function () {
    await runner.stop()
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('home page has expected title', async function () {
    await getRootAndWait()
    const title = await webdriver.getTitle()
    expect(title).to.equal('WackyTracky')
  })
})
