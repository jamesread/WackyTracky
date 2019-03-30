#!/usr/bin/python3

from os import getenv
from neo4j.v1 import GraphDatabase
from neo4j.exceptions import ServiceUnavailable
import logging
import configparser

class Wrapper:
  def __init__(self, username = None, password = None):
    if username == None or password == None:
        username, password = self.loadAuthCredentials()

    uri = "bolt://{}:{}@localhost".format(username, password)

    try:
      self.conn = GraphDatabase.driver(uri, auth = (username, password))
      self.session = self.conn.session()
    except ServiceUnavailable as e:
      raise Exception(str(e))


  def loadAuthCredentials(self):
    config = configparser.ConfigParser()
    config.read("/etc/wacky-tracky/server.cfg")

    return config["server"]["dbUser"], config["server"]["dbPassword"]

  def createUser(self, username):
    results = self.session.run("CREATE (u:User {username: {username}})", username = username)

  def deleteUser(self, userid):
    results = self.session.run("MATCH (u:User) WHERE id(u) = {userid} DELETE u ", userid = userid)

  def getUsers(self):
    results = self.session.run("MATCH (u:User) OPTIONAL MATCH (u)-[]->(l:List) OPTIONAL MATCH (u)-[]->(l:List)-[]->(i2:Item) RETURN u, count(l) AS listCount, count(i2) AS itemCount ORDER BY id(u) ");

    return results;

  def setTaskContent(self, itemId, content):
    results = self.session.run("MATCH (i:Item) WHERE id(i) = {itemId} SET i.content = {content} ", itemId = itemId, content = content)

  def createList(self, username, title):
    results = self.session.run("MATCH (u:User) WHERE u.username = {username} CREATE (u)-[:owns]->(l:List {title: {title}}) RETURN id(l)", title = title, username = username);

    return results;

  def getSingleList(self, username, listId):
    results = self.session.run("MATCH (u:User)-[]->(l:List) OPTIONAL MATCH (l)-[]->(subLists:List) OPTIONAL MATCH (l)-->(i:Item) WITH u, l, count(subLists) AS countSubLists, count(i) AS countItems WHERE u.username = {username} AND id(l) = {listId} RETURN l, countSubLists, countItems ORDER BY l.title", username = username, listId = listId).single();

    return results;

  def getListByTitle(self, username, listTitle):
    results = self.session.run("MATCH (u:User)-[]->(l:List) WHERE u.username = {username} AND l.title = {listTitle} RETURN l ORDER BY l.title", username = username, listTitle = listTitle);

    return results;

  def getLists(self, username):
    results = self.session.run("MATCH (u:User)-[]->(l:List) OPTIONAL MATCH (l)-->(subLists:List) OPTIONAL MATCH (l)-->(i:Item) WITH u, l, count(subLists) AS countSubLists, count(i) AS countItems WHERE u.username = {username} RETURN l, countSubLists, countItems ORDER BY l.title", username = username);

    return results;

  def getTags(self, username):
    results = self.session.run("MATCH (u:User)-[]->(t:Tag) OPTIONAL MATCH (t)-[]->(tv:TagValue) WHERE u.username = {username} RETURN t, tv ", username = username);

    return results;

  def setItemParent(self, username, item, newParent):
    results = self.session.run("MATCH (u:User)-[]->(i:Item) CREATE ")

    return results

  def createTagValue(self, tagId)
    results = self.session.run("MATCH (t:Tag) WHERE id(t) = {tagId} CREATE (tv)-[:type]->(tv:TagValue) ", username = username, title = title);

  def createTag(self, username, title):
    results = self.session.run("MATCH (u:User) WHERE u.username = {username} CREATE (u)-[:owns]->(t:Tag {title: {title}})-[:type]->(tv:TagValue {textualValue:'default'}) ", username = username, title = title);

  def createListItem(self, listId, content):
    return self.session.run("MATCH (l:List) WHERE id(l) = {listId} CREATE (l)-[:owns]->(i:Item {content: {content}}) WITH i, 0 as countItems RETURN i, countItems", listId = listId, content = content).single()

  def createOrFindListItem(self, listId, content, externalId):
    listItem = self.session.run("MATCH (i:ExternalItem) WHERE i.externalId = {externalId} RETURN i", externalId = externalId).single()

    if listItem is None:
      return self.createListItem(listId, content)
    else:
      return listItem
  
  def createSubItem(self, itemId, content):
    return self.session.run("MATCH (pi:Item) WHERE id(pi) = {itemId} CREATE (pi)-[:owns]->(i:Item {content: {content}}) WITH i, 0 as countItems RETURN i, countItems", itemId = itemId, content = content).single()

  def getItemsFromList(self, listId, sort = None):
    if sort not in [ "content", "dueDate" ]:
      sort = "content"

    results = self.session.run("MATCH (l:List)-[]->(i:Item) OPTIONAL MATCH (i)-->(tv:TagValue) OPTIONAL MATCH (i)-->(subItem:Item) OPTIONAL MATCH (externalItem:ExternalItem) WHERE i = externalItem WITH l, i, count(tv) AS countTagValues, count(subItem) AS countItems, externalItem WHERE id(l) = {listId} WITH i, countTagValues, countItems, externalItem RETURN i, countTagValues, countItems, externalItem ORDER BY i." + sort, listId = listId);

    return results

  def deleteTag(self, itemId):
    results = self.session.run("MATCH (t:Tag)-[tr]-() WHERE id(t) = {tagId} DELETE tr, t", tagId = itemId)
    return results

  def getItemTags(self, itemId):
    ret = []

    print("Get Item Tags for item: " + str(itemId))

    results = self.session.run("MATCH (i)-->(tv:TagValue)<--(t:Tag) WHERE id(i) = {itemId} RETURN tv, t", itemId = itemId)

    for row in results:
        numericValue = 0
        if row['tv']['numericValue'] not in ("", None):
            numericValue = int(row['tv']['numericValue'])

        ret.append({
            "tagId": row['t'].id,
            "title": row['t']['title'],
            "numericValue": numericValue,
            "textualValue": row['tv']['textualValue'],
            "backgroundColor": row['tv']['backgroundColor']
        })

    return ret

  def addItemLabel(self, itemId, label):
    self.session.run("MATCH (i) WHERE id(i) = {id} SET i:" + label + "  RETURN i", id = itemId)

  def setItemProperty(self, itemId, key, val):
    self.session.run("MATCH (i) WHERE id(i) = {id} SET i." + key + " = {value}  RETURN i", id = itemId, value = val)

  def getSubItems(self, parentId, sort = None):
    if sort not in [ "content", "dueDate" ]:
      sort = "content"

    results = self.session.run("MATCH (p:Item)-[]->(i:Item) WHERE id(p) = {parentId} OPTIONAL MATCH (i)-->(subItem:Item) WITH i, count(subItem) as countItems RETURN i, countItems ORDER BY i." + sort, parentId = parentId);

    return results

  def deleteTask(self, itemId):
    results = self.session.run("MATCH (i:Item) WHERE id(i) = {itemId} OPTIONAL MATCH (i)<-[r]-() OPTIONAL MATCH (i)-[linkTagged]->(tv:TagValue) DELETE i,r, linkTagged, tv", itemId = itemId)

  def updateList(self, listId, title, sort):
    results = self.session.run("MATCH (l:List) WHERE id(l) = {listId} SET l.title = {title}, l.sort = {sort} ", listId = listId, title = title, sort = sort);

  def setDueDate(self, itemId, dueDate):
    results = self.session.run("MATCH (i:Item) WHERE id(i) = {itemId} SET i.dueDate = {dueDate} ", itemId = itemId, dueDate = dueDate);

  def updateTag(self, itemId, title, shortTitle, backgroundColor, textualValue, numericValue, tagValueId):
    try:
        numericValue = int(numericValue)
    except:
        numericValue = 1

    cql =  "MATCH (t:Tag) WHERE id(t) = {itemId} MATCH (tv:TagValue) WHERE id(tv) = {tagValueId} SET t.title = {title}, t.shortTitle = {shortTitle}, tv.backgroundColor = {backgroundColor}, tv.textualValue = {textualValue}, tv.numericValue = {numericValue} RETURN t, tv"

    results = self.session.run(cql, itemId = itemId, title = title, shortTitle = shortTitle, backgroundColor = backgroundColor, textualValue = textualValue, numericValue = numericValue, tagValueId = tagValueId);


    for result in results:
      print("xxxxxxxx update tag and tv !")

      tag = result['t']
      tv = result['tv']

      return {
        "id": tag.id,
        "title": tag['title'],
        "shortTitle": tag['shortTitle'],
        "tagValueId": tv.id,
        "numericValue": tv["numericValue"],
        "textualValue": tv["textualValue"],
        "backgroundColor": tv['backgroundColor']
      }

    return None

  def deleteList(self, itemId):
    results = self.session.run("MATCH (l:List) WHERE id(l) = {listId} OPTIONAL MATCH (l)<-[userLink]-() DELETE l, userLink", listId = itemId);

  def tag(self, itemId, tagId, tagValueId):
    results = self.session.run("MATCH (i:Item), (t:Tag)-->(tv:TagValue) WHERE id(i) = {itemId} AND id(t) = {tagId} AND id(tv) = {tagValueId} CREATE (i)-[:tagged]->(tv) ", tagId = tagId, itemId = itemId, tagValueId = tagValueId);

  def untag(self, itemId, tagId, tagValueId):
    results = self.session.run("MATCH (i:Item)-[link:tagged]->(tv:TagValue) WHERE id(i) = {itemId} AND id(tv) = {tagValueId} DELETE link ", itemId = itemId, tagValueId = tagValueId);

  def hasItemGotTag(self, itemId, tagValueId):
    results = self.session.run("MATCH (i:Item)-[r]->(tv:TagValue) WHERE id(i) = {itemId} AND id(tv) = {tagValueId} RETURN r", itemId = itemId, tagValueId = tagValueId);

    tagCount = 0;

    for r in results:
      tagCount += 1

    return tagCount > 0

  def register(self, username, hashedPassword, email):  
    results = self.session.run("CREATE (u:User {username: {username}, password: {password}, email: {email}}) ", username = usenrame, password = hashedPassword, email = email)

    return;

  def changePassword(self, username, hashedPassword):
    results = self.session.run("MATCH (u:User) WHERE u.username = {username} SET u.password = {hashedPassword} RETURN u", username = username, hashedPassword = hashedPassword);

    return results


  def getUser(self, username = None):
    if username == None:
      username = username

    results = self.session.run("MATCH (u:User) WHERE u.username = {username} RETURN u LIMIT 1", username = username);
    results = results.values()

    if len(results) == 0:
      return [None, None]
    else: 
      for row in results:
        user = row[0]

        return [{
          "username": user['username'],
        }, user['password']]

def fromArgs(args):
    return Wrapper(args.dbUser, args.dbPassword)

import __main__ as main
if not hasattr(main, '__file__'):
  global wtapi
  wtapi = Wrapper()
  print("WT API reloaded")
