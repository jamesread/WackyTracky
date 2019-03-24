#!/usr/bin/python3

import __init__
import wrapper

api = wrapper.Wrapper()

item = api.createListItem(41, "imported")['i']

print("created: " + str(item.id))
api.addItemLabel(item.id, "ExternalItem")
api.setItemProperty(item.id, "source", "sample")
api.setItemProperty(item.id, "url", "http://example.com")
api.setItemProperty(item.id, "icon", "sample.png")
