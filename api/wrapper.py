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
    results = self.session.run("MATCH (u:User)-[]->(t:Tag) WHERE u.username = {username} RETURN t ", username = username);

    return results;

  def setItemParent(self, username, item, newParent):
    results = self.session.run("MATCH (u:User)-[]->(i:Item) CREATE ")

    return results

  def createTag(self, username, title):
    results = self.session.run("MATCH (u:User) WHERE u.username = {username} CREATE (u)-[:owns]->(t:Tag {title: {title}}) ", username = username, title = title);

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

    results = self.session.run("MATCH (l:List)-[]->(i:Item) OPTIONAL MATCH (i)-->(subItem:Item) OPTIONAL MATCH (externalItem:ExternalItem) WHERE i = externalItem WITH l, i, count(subItem) AS countItems, externalItem WHERE id(l) = {listId} WITH i, countItems, externalItem RETURN i, countItems, externalItem ORDER BY i." + sort, listId = listId);

    return results

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
    results = self.session.run("MATCH (i:Item) WHERE id(i) = {itemId} OPTIONAL MATCH (i)<-[r]-() OPTIONAL MATCH (i)-[linkTagged:tagged]->(tag:Tag) DELETE i,r, linkTagged, tag", itemId = itemId)

  def updateList(self, listId, title, sort, timeline):
    results = self.session.run("MATCH (l:List) WHERE id(l) = {listId} SET l.title = {title}, l.sort = {sort}, l.timeline = {timeline} ", listId = listId, title = title, sort = sort, timeline = timeline);

  def setDueDate(self, itemId, dueDate):
    results = self.session.run("MATCH (i:Item) WHERE id(i) = {itemId} SET i.dueDate = {dueDate} ", itemId = itemId, dueDate = dueDate);

  def updateTag(self, itemId, title, shortTitle, backgroundColor):
    results = self.session.run("MATCH (t:Tag) WHERE id(t) = {itemId} SET t.title = {title}, t.shortTitle = {shortTitle}, t.backgroundColor = {backgroundColor} RETURN t", itemId = itemId, title = title, shortTitle = shortTitle, backgroundColor = backgroundColor);

    for result in results:
      tag = result[0]

      return {
        "id": tag.id,
        "title": tag['title'],
        "shortTitle": tag['shortTitle'],
        "backgroundColor": tag['backgroundColor']
      }

    return None

  def deleteList(self, itemId):
    results = self.session.run("MATCH (l:List) WHERE id(l) = {listId} OPTIONAL MATCH (l)<-[userLink]-() DELETE l, userLink", listId = itemId);

  def tag(self, itemId, tagId):
    results = self.session.run("MATCH (i:Item), (t:Tag) WHERE id(i) = {itemId} AND id(t) = {tagId} CREATE UNIQUE (i)-[:tagged]->(t) ", tagId = tagId, itemId = itemId);

  def untag(self, itemId, tagId):
    results = self.session.run("MATCH (i:Item)-[link:tagged]->(t:Tag) WHERE id(i) = {itemId} AND id(t) = {tagId} DELETE link ", itemId = itemId, tagId = tagId);

  def hasItemGotTag(self, itemId, tagId):
    results = self.session.run("MATCH (i:Item)-[r]->(t:Tag) WHERE id(i) = {itemId} AND id(t) = {tagId} RETURN r", itemId = itemId, tagId = tagId);

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
