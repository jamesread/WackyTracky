#!/usr/bin/python3

import __init__
import wrapper
import json
import configargparse

import requests

parser = configargparse.ArgumentParser(default_config_files=["/etc/wacky-tracky/import-trello.cfg"])
parser.add_argument("--apiKey", required=True)
parser.add_argument("--apiSecret", required=True)
parser.add_argument('--findLists', action = 'store_true')
parser.add_argument("--board", required = True)
parser.add_argument("--listIds", nargs = '+', default = [])
parser.add_argument("--importToList", type=int, required = True)
args = parser.parse_args();

authString = "key=%s&token=%s" % (args.apiKey, args.apiSecret)

apiBase = "https://api.trello.com/"

class TrelloObject:
  name = ""
  id = ""

class TrelloBoard(TrelloObject):
  pass

class TrelloList(TrelloObject):
  pass

class TrelloCard(TrelloObject):
  pass

class TrelloApiLayer:
  def trelloApiCall(self, method):
    if not method.endswith("?"):
      method += "?"

    url = apiBase + method + authString

    return requests.get(url)

  def getBoards(self):
    return self.get("1/members/me/boards/")

  def getLists(self, boardId):
    return self.get("1/boards/" + boardId + "/lists")

  def getCards(self, boardId):
    return self.get("1/boards/" + boardId + "/cards")

  def get(self, url):
    res = self.trelloApiCall(url)

    if res.status_code == 200:
      return res.json()
    else:
      return None


class PojoTransformer:
  api = TrelloApiLayer()

  def getBoards(self):
    res = self.api.getBoards()
    
    ret = []
    for rec in res:
      board = TrelloBoard()
      board.id = rec['id']
      board.name = rec['name']

      ret.append(board);

    return ret

  def getLists(self, boardId):
    res = self.api.getLists(boardId)

    ret = []
    for rec in res:
      l = TrelloList()
      l.id = rec['id']
      l.name = rec['name']

      ret.append(l)

    return ret



  def getCards(self, boardId, listIds):
    res = self.api.getCards(boardId)

    ret = []
    for rec in res:
      if rec['idList'] not in listIds: continue

      card = TrelloCard()
      card.id = rec['id']
      card.name = rec['name']
      card.url = rec['url']

      ret.append(card)

    return ret

transformer = PojoTransformer()

if args.findLists:
  for board in transformer.getBoards():
    print(board.id, board.name)
    for l in transformer.getLists(board.id):
      print("--", l.id, l.name)

api = wrapper.Wrapper()
  
for card in transformer.getCards(args.board, args.listIds):
  item = api.createOrFindListItem(args.importToList, card.name, card.id)['i']

  api.addItemLabel(item.id, "ExternalItem")
  api.setItemProperty(item.id, "content", card.name) # because we might be updating
  api.setItemProperty(item.id, "externalId", card.id)
  api.setItemProperty(item.id, "source", "trello")
  api.setItemProperty(item.id, "url", card.url)
  api.setItemProperty(item.id, "icon", "trello.png")

  print("Created item", card.name)

