#!/usr/bin/env python3 

import wrapper

def test_get_lists():
  api = wrapper.fromParams()

  assert api != None

  listResults = api.getLists("testing")
  
  assert listResults != None


