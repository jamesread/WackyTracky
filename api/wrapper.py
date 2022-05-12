from wrapper_neo4j import WrapperNeo4j
from wrapper_sql_mysql import WrapperMySQL

def fromParams(username = None, password = None, server = None):
  driver = "mysql"

  if username == None or password == None:
    cfg = commonArgumentParser.getParsed();
    username = cfg.dbUser
    password = cfg.dbPassword
    server = cfg.dbServer
    driver = cfg.dbDriver

  print("Driver = ", driver)

  if driver == "neo4j":
    return WrapperNeo4j(username, password, server);

  if driver == "mysql":
    return WrapperMySQL(username, password, server);


 
def fromArgs(args):
  return fromParams(args.dbUser, args.dbPassword, args.dbServer)

class Wrapper:
<<<<<<< HEAD
  def createUser(self, username):
    raise NotImplementedError

  def deleteUser(self, userid):
    raise NotImplementedError

  def getUsers(self):
    raise NotImplementedError

  def setTaskContent(self, itemId, content):
    raise NotImplementedError

  def createList(self, username, title):
    raise NotImplementedError

  def getSingleList(self, username, listId):
    raise NotImplementedError

  def getListByTitle(self, username, listTitle):
    raise NotImplementedError

  def getLists(self, username):
    raise NotImplementedError

  def getTags(self, username):
    raise NotImplementedError

  def setItemParent(self, username, item, newParent):
    raise NotImplementedError

  def createTagValue(self, tagId):
    raise NotImplementedError

  def createTag(self, username, title):
    raise NotImplementedError

  def createListItem(self, listId, content):
    raise NotImplementedError

  def createOrFindListItem(self, listId, content, externalId):
    raise NotImplementedError
=======
  def __init__(self, username = None, password = None, server = None):
    if username == None or password == None:
        username, password, server = self.loadAuthCredentials()

    self.server = server
    self.username = username;
    self.password = password
    self.conn = None


  def getSession(self):
    if self.conn == None or self.session.closed():
      try:
        uri = "bolt://{}:{}@{}".format(self.username, self.password, self.server)
        print("Connecting to " + uri)

        self.conn = GraphDatabase.driver(uri, auth = (self.username, self.password))
        self.session = self.conn.session()
      except ServiceUnavailable as e:
        raise Exception(str(e))

    return self.session

  def loadAuthCredentials(self):
    config = configparser.ConfigParser()
    config.read("/etc/wacky-tracky/server.cfg")

    return config["server"]["dbUser"], config["server"]["dbPassword"], config["server"]["dbServer"]

  def createUser(self, username):
    results = self.getSession().run("CREATE (u:User {username: {username}})", username = username)

  def deleteUser(self, userid):
    results = self.getSession().run("MATCH (u:User) WHERE id(u) = {userid} DELETE u ", userid = userid)

  def getUsers(self):
    results = self.getSession().run("MATCH (u:User) OPTIONAL MATCH (u)-[]->(l:List) OPTIONAL MATCH (u)-[]->(l:List)-[]->(i2:Item) RETURN u, count(l) AS listCount, count(i2) AS itemCount ORDER BY id(u) ");

    return results;

  def setTaskContent(self, itemId, content):
    results = self.getSession().run("MATCH (i:Item) WHERE id(i) = {itemId} SET i.content = {content} ", itemId = itemId, content = content)

  def createList(self, username, title):
    results = self.getSession().run("MATCH (u:User) WHERE u.username = {username} CREATE (u)-[:owns]->(l:List {title: {title}}) RETURN id(l)", title = title, username = username);

    return results;

  def getSingleList(self, username, listId):
    results = self.getSession().run("MATCH (u:User)-[]->(l:List) OPTIONAL MATCH (l)-[]->(subLists:List) OPTIONAL MATCH (l)-->(i:Item) WITH u, l, count(subLists) AS countSubLists, count(i) AS countItems WHERE u.username = {username} AND id(l) = {listId} RETURN l, countSubLists, countItems ORDER BY l.title", username = username, listId = listId).single();

    return results;

  def getListByTitle(self, username, listTitle):
    results = self.getSession().run("MATCH (u:User)-[]->(l:List) WHERE u.username = {username} AND l.title = {listTitle} RETURN l ORDER BY l.title", username = username, listTitle = listTitle);

    return results;

  def getLists(self, username):
    results = self.getSession().run("MATCH (u:User)-[]->(l:List) OPTIONAL MATCH (l)-->(subLists:List) OPTIONAL MATCH (l)-->(i:Item) WITH u, l, count(subLists) AS countSubLists, count(i) AS countItems WHERE u.username = {username} RETURN l, countSubLists, countItems ORDER BY l.title", username = username);

    return results;

  def getTags(self, username):
    results = self.getSession().run("MATCH (u:User)-[]->(t:Tag) OPTIONAL MATCH (t)-[]->(tv:TagValue) WHERE u.username = {username} RETURN t, tv ORDER BY t.title, tv.numericValue", username = username);

    return results;

  def setItemParent(self, username, item, newParent):
    results = self.getSession().run("MATCH (u:User)-[]->(i:Item) CREATE ")

    return results

  def createTagValue(self, tagId):
    results = self.getSession().run("MATCH (t:Tag) WHERE id(t) = {tagId} CREATE (t)-[:type]->(tv:TagValue {textualValue:'default'}) ", tagId = tagId);

    return results

  def createTag(self, username, title):
    results = self.getSession().run("MATCH (u:User) WHERE u.username = {username} CREATE (u)-[:owns]->(t:Tag {title: {title}})-[:type]->(tv:TagValue {textualValue:'default'}) ", username = username, title = title);

  def createListItem(self, listId, content):
    return self.getSession().run("MATCH (l:List) WHERE id(l) = {listId} CREATE (l)-[:owns]->(i:Item {content: {content}}) WITH i, 0 as countItems RETURN i, countItems", listId = listId, content = content).single()

  def createOrFindListItem(self, listId, content, externalId):
    listItem = self.getSession().run("MATCH (i:ExternalItem) WHERE i.externalId = {externalId} RETURN i", externalId = externalId).single()

  def createSubItem(self, itemId, content):
    raise NotImplementedError

  def getItemsFromList(self, listId, sort = None):
    raise NotImplementedError
    
  def deleteTag(self, itemId):
    raise NotImplementedError
    
  def getItemTags(self, itemId):
    raise NotImplementedError

  def addItemLabel(self, itemId, label):
    raise NotImplementedError

  def setItemProperty(self, itemId, key, val):
    raise NotImplementedError

  def getSubItems(self, parentId, sort = None):
    raise NotImplementedError

  def deleteTask(self, itemId):
    raise NotImplementedError

  def updateList(self, listId, title, sort):
    raise NotImplementedError

  def setDueDate(self, itemId, dueDate):
    raise NotImplementedError

  def updateTag(self, itemId, title, shortTitle, backgroundColor, textualValue, numericValue, tagValueId):
    raise NotImplementedError

  def deleteList(self, itemId):
    raise NotImplementedError

  def tag(self, itemId, tagId, tagValueId):
    raise NotImplementedError

  def untag(self, itemId, tagId, tagValueId):
    raise NotImplementedError

  def hasItemGotTag(self, itemId, tagValueId):
    raise NotImplementedError

  def register(self, username, hashedPassword, email):  
    raise NotImplementedError

  def changePassword(self, username, hashedPassword):
    return self.getSession().run("MATCH (pi:Item) WHERE id(pi) = {itemId} CREATE (pi)-[:owns]->(i:Item {content: {content}}) WITH i, 0 as countItems RETURN i, countItems", itemId = itemId, content = content).single()

  def getItemsFromList(self, listId, sort = None):
    if sort not in [ "content", "dueDate", "id" ]:
      sort = "content"

    if sort == "id":
        sort = "id(i)"
    else:
        sort = "i." + sort
 
    print("getItemsFromList: " + str(listId))

    results = self.getSession().run("MATCH (l:List)-[]->(i:Item) OPTIONAL MATCH (i)-->(tv:TagValue) OPTIONAL MATCH (i)-->(subItem:Item) OPTIONAL MATCH (externalItem:ExternalItem) WHERE i = externalItem WITH l, i, count(tv) AS countTagValues, count(subItem) AS countItems, externalItem WHERE id(l) = {listId} WITH i, countTagValues, countItems, externalItem RETURN i, countTagValues, countItems, externalItem ORDER BY " + sort, listId = listId);

    return results

  def deleteTag(self, itemId):
    results = self.getSession().run("MATCH (t:Tag)-[tr]-() WHERE id(t) = {tagId} DELETE tr, t", tagId = itemId)
    return results

  def getItemTags(self, itemId):
    ret = []

    print("Get Item Tags for item: " + str(itemId))

    results = self.getSession().run("MATCH (i)-->(tv:TagValue)<--(t:Tag) WHERE id(i) = {itemId} RETURN tv, t ORDER BY t.title", itemId = itemId)

    for row in results:
        numericValue = 0
        if row['tv']['numericValue'] not in ("", None):
            numericValue = int(row['tv']['numericValue'])

        ret.append({
            "id": row['t'].id,
            "tagValueId": row['tv'].id,
            "title": row['t']['title'],
            "numericValue": numericValue,
            "textualValue": row['tv']['textualValue'],
            "backgroundColor": row['tv']['backgroundColor']
        })

    return ret

  def addItemLabel(self, itemId, label):
    self.getSession().run("MATCH (i) WHERE id(i) = {id} SET i:" + label + "  RETURN i", id = itemId)

  def setItemProperty(self, itemId, key, val):
    self.getSession().run("MATCH (i) WHERE id(i) = {id} SET i." + key + " = {value}  RETURN i", id = itemId, value = val)

  def getSubItems(self, parentId, sort = None):
    if sort not in [ "content", "dueDate" ]:
      sort = "content"

    print("Get Sub Items, parentId: " + str(parentId))

    results = self.getSession().run("MATCH (p:Item)-[]->(i:Item) WHERE id(p) = {parentId} OPTIONAL MATCH (i)-->(subItem:Item) WITH i, count(subItem) as countItems RETURN i, countItems ORDER BY i." + sort, parentId = parentId);

    return results

  def deleteTask(self, itemId):
    results = self.getSession().run("MATCH (i:Item) WHERE id(i) = {itemId} OPTIONAL MATCH (i)<-[r]-() OPTIONAL MATCH (i)-[linkTagged]->(tv:TagValue) DELETE i,r, linkTagged, tv", itemId = itemId)

  def updateList(self, listId, title, sort):
    results = self.getSession().run("MATCH (l:List) WHERE id(l) = {listId} SET l.title = {title}, l.sort = {sort} ", listId = listId, title = title, sort = sort);

  def setDueDate(self, itemId, dueDate):
    results = self.getSession().run("MATCH (i:Item) WHERE id(i) = {itemId} SET i.dueDate = {dueDate} ", itemId = itemId, dueDate = dueDate);

  def updateTag(self, itemId, title, shortTitle, backgroundColor, textualValue, numericValue, tagValueId):
    try:
        numericValue = int(numericValue)
    except:
        numericValue = 1

    cql =  "MATCH (t:Tag) WHERE id(t) = {itemId} MATCH (tv:TagValue) WHERE id(tv) = {tagValueId} SET t.title = {title}, t.shortTitle = {shortTitle}, tv.backgroundColor = {backgroundColor}, tv.textualValue = {textualValue}, tv.numericValue = {numericValue} RETURN t, tv"

    results = self.getSession().run(cql, itemId = itemId, title = title, shortTitle = shortTitle, backgroundColor = backgroundColor, textualValue = textualValue, numericValue = numericValue, tagValueId = tagValueId);


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
    results = self.getSession().run("MATCH (l:List) WHERE id(l) = {listId} OPTIONAL MATCH (l)<-[userLink]-() DELETE l, userLink", listId = itemId);

  def tag(self, itemId, tagId, tagValueId):
    results = self.getSession().run("MATCH (i:Item), (t:Tag)-->(tv:TagValue) WHERE id(i) = {itemId} AND id(t) = {tagId} AND id(tv) = {tagValueId} CREATE (i)-[:tagged]->(tv) ", tagId = tagId, itemId = itemId, tagValueId = tagValueId);

  def untag(self, itemId, tagId, tagValueId):
    results = self.getSession().run("MATCH (i:Item)-[link:tagged]->(tv:TagValue)<-[]-(t:Tag) WHERE id(i) = {itemId} AND id(t) = {tagId} DELETE link ", itemId = itemId, tagId = tagId);

  def hasItemGotTag(self, itemId, tagValueId):
    results = self.getSession().run("MATCH (i:Item)-[r]->(t:TagValue) WHERE id(i) = {itemId} AND id(t) = {tagValueId} RETURN r", itemId = itemId, tagValueId = tagValueId);
    results = results.values()

    return len(results) > 0

  def register(self, username, hashedPassword, email):  
    results = self.getSession().run("CREATE (u:User {username: {username}, password: {password}, email: {email}}) ", username = usenrame, password = hashedPassword, email = email)

    return;

  def changePassword(self, username, hashedPassword):
    results = self.getSession().run("MATCH (u:User) WHERE u.username = {username} SET u.password = {hashedPassword} RETURN u", username = username, hashedPassword = hashedPassword);

    return results

  def getUser(self, username = None):
    results = self.getSession().run("MATCH (u:User) WHERE u.username = {username} RETURN u LIMIT 1", username = username);
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
    return Wrapper(args.dbUser, args.dbPassword, args.dbServer)

import __main__ as main
if not hasattr(main, '__file__'):
  global wtapi
  wtapi = Wrapper()
  
  print("WT API reloaded")


