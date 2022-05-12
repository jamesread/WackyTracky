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
    raise NotImplementedError

  def getUser(self, username = None):
    raise NotImplementedError

