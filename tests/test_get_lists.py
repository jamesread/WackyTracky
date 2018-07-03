import wrapper


def test_get_lists():
  api = wrapper.Wrapper()

  assert api != None

  listResults = api.getLists("testing")
  
  assert listResults != None


