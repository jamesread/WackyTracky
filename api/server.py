#!/usr/bin/python3

import cherrypy
from cherrypy import _cperror
from cherrypy._cperror import HTTPError
from cherrypy.lib import sessions
import wrapper
import json
import random
import os
from sys import exc_info
import io
import csv

import commonArgumentParser

JSON_OK = { "message": "ok"}

parser = commonArgumentParser.getNew();
parser.add_argument("--port", default = 8082, type = int)
parser.add_argument("--wallpaperdir", default = "/var/www/html/wallpapers/")
parser.add_argument("--background", action = 'store_true')
parser.add_argument("--corsDomain", default = "*")
args = parser.parse_args();

class HttpQueryArgChecker:
  def __init__(self, args):
    self.args = args;

  def requireArg(self, arg):
    if arg not in self.args:
      raise cherrypy.HTTPError(403, "Argument not provided:" + arg)

    return self

class Api(object):
  wrapper = wrapper.Wrapper(args.dbUser, args.dbPassword, args.dbServer)

  @cherrypy.expose
  def tag(self, *path, **args):
    if self.wrapper.hasItemGotTag(int(args['item']), int(args['tagValueId'])):
        self.wrapper.untag(int(args['item']), int(args['tag']), int(args['tagValueId']));
    else:
        self.wrapper.untag(int(args['item']), int(args['tag']), int(args['tagValueId']));
        self.wrapper.tag(int(args['item']), int(args['tag']), int(args['tagValueId']));

    return self.outputJson(JSON_OK);

  @cherrypy.expose
  def listUpdate(self, *path, **args):
    self.wrapper.updateList(int(args['id']), args['title'], args['sort'])

    return self.outputJson(JSON_OK);

  @cherrypy.expose
  def setDueDate(self, *path, **args):
    self.wrapper.setDueDate(int(args['item']), args['dueDate'])

    return self.outputJson(JSON_OK);

  @cherrypy.expose
  def default(self, *args, **kwargs):
    return "wt API"

  def outputJson(self, structure, download=False, downloadFilename = "output.json"):
    if download:
      cherrypy.response.headers['Content-Disposition'] = 'attachment; filename="' + downloadFilename + '" '
      
    cherrypy.response.headers['Content-Type'] = 'application/json'

    return json.dumps(structure, indent = 4).encode("utf8");

  @cherrypy.expose
  def createTag(self, *path, **args):
    self.wrapper.createTag(self.getUsername(), args['title'])

    return self.outputJson(JSON_OK);

  @cherrypy.expose
  def addTagValue(self, *path, **args):
    self.wrapper.createTagValue(int(args['tagId']))

    return self.outputJson(JSON_OK);

  @cherrypy.expose
  def listDownload(self, *path, **args):
    items = self.wrapper.getItemsFromList(int(args['id']));

    ret = []
    for row in items: 
      singleItem = row

      ret.append(self.normalizeItem(singleItem))

    if args['format'] == "json":
      return self.outputJson(ret, download = True, downloadFilename = "list" + args['id'] + ".json");
    elif args['format'] == "csv": 
      retStr = io.StringIO()      
      writer = csv.writer(retStr);
      writer.writerow(["ID", "Content", "Numeric Product"])

      for item in ret:
        writer.writerow([item['id'], item['content'], item['tagNumericProduct']])

      cherrypy.response.headers['Content-Disposition'] = 'attachment; filename="list' + args['id'] + '.csv" '
      cherrypy.response.headers['Content-Type'] = 'text/csv'

      return retStr.getvalue();
    else:
      retStr = ""

      for item in ret:
        retStr += item['content'] + "\n"

      cherrypy.response.headers['Content-Disposition'] = 'attachment; filename="list' + args['id'] + '.txt" '
      cherrypy.response.headers['Content-Type'] = 'text/plain'

      return retStr

  @cherrypy.expose
  def listTags(self, *path, **args):
    tags = self.wrapper.getTags(self.getUsername());

    ret = []

    for row in tags:
      singleTag = row['t']
      tagValue = row['tv']

      ret.append({
        "id": singleTag.id,
        "title": singleTag['title'],
        "shortTitle": singleTag['shortTitle'],
        "tagValueId": tagValue.id,
        "numericValue": tagValue["numericValue"],
        "textualValue": tagValue["textualValue"],
        "backgroundColor": tagValue['backgroundColor']
      });

    return self.outputJson(ret);

  @cherrypy.expose
  def getList(self, *path, **args):
    self.checkLoggedIn();
    HttpQueryArgChecker(args).requireArg('listId');

    singleList = self.wrapper.getSingleList(self.getUsername(), int(args['listId']));

    if singleList == None:
      return self.outputJsonError(404, "List not found")

    structList = self.normalizeList(singleList);

    return self.outputJson(structList);

  @cherrypy.expose
  def getListByTitle(self, *path, **args):
    self.checkLoggedIn();

    HttpQueryArgChecker(args).requireArg('listTitle');

    l = self.wrapper.getListByTitle(self.getUsername(), args['listTitle'])

    if len(l) == 0:
      return self.outputJsonError(404, "List not found by title", uniqueType = "list-not-found");

    singleList = l[0][0]
    structList = self.normalizeList(singleList)
    
    return self.outputJson(structList);

  @cherrypy.expose
  def listLists(self, *path, **args):
    listResults = self.wrapper.getLists(self.getUsername());

    ret = []

    for listRecord in listResults:
      ret.append(self.normalizeList(listRecord))

    return self.outputJson(ret)

  def normalizeList(self, listRecord):
    singleList = listRecord['l']

    return {  
      "id": singleList.id,
      "title": singleList['title'],
      "sort": singleList['sort'],
      "timeline": singleList['timeline'],
      "countSubLists": listRecord['countSubLists'],
      "countItems": listRecord['countItems']
    }


  @cherrypy.expose
  def createList(self, *path, **args):
    createdList = self.wrapper.createList(self.getUsername(), args["title"]);

    for row in createdList:
      newListId = row[0]

      return self.outputJson({"newListId": newListId})

    return self.outputJsonError(404, "no list created");

  @cherrypy.expose
  def setItemParent(self, *path, **args):
    self.wrapper.setItemParent(self.getUsername(), args['item'], args['newParent'])

    return self.outputJson(JSON_OK);

  @cherrypy.expose
  def createTask(self, *path, **args):
    if (args['parentType'] == "list"):
      listRecord = self.wrapper.getSingleList(self.getUsername(), int(args['parentId']));

      print("lr = ", listRecord)

#      if listRecord == None:
#        return self.outputJsonError(404, "Cannot get the owning list in which to create a task.", uniqueType = "create-on-nonexistant-list")

      structList = self.normalizeList(listRecord);

      if (structList['countItems'] > 50):
        raise self.outputJsonError(403, "List is too big!")

      createdItems = self.wrapper.createListItem(int(args['parentId']), args['content'])
    else:
      createdItems = self.wrapper.createSubItem(int(args['parentId']), args['content'])

    return self.outputJson(self.normalizeItem(createdItems));

  def normalizeItem(self, singleItemRecord):
    item = singleItemRecord['i']
    url = ""
    source = ""
    icon = ""
    tags = []
    tagNumericProduct = 1

    if "url" in item:
      url = item["url"]

    if "source" in item:
      source = item["source"]

    if "icon" in item:
      icon = item["icon"]

    if 'countTagValues' in singleItemRecord.keys() and singleItemRecord['countTagValues'] > 0:
        tags = self.wrapper.getItemTags(item.id)

        for tag in tags: 
            if tag['numericValue'] > 0:
                tagNumericProduct *= tag['numericValue']

    return {
      "hasChildren": singleItemRecord['countItems'] > 0,
      "content": item['content'],
      "tags": tags,
      "tagNumericProduct": tagNumericProduct,
      "dueDate": item['dueDate'],
      "id": item.id,
      "url": url,
      "source": source,
      "icon": icon
    }

  @cherrypy.expose
  def listTasks(self, *path, **args):
    self.checkLoggedIn();
    
    databaseSorts = ["default", "content", "id"]
    pythonSorts = ["tagNumericProduct"]

    sortBy = args['sort']

    if sortBy not in databaseSorts:
        sortBy = "content"

    if "task" in args:
      print("listTasks, getSubItems")
      items = self.wrapper.getSubItems(int(args['task']), sortBy)
    else:
      print("listTasks, getItemsFromList")
      items = self.wrapper.getItemsFromList(int(args['list']), sortBy)

    ret = []
    for itemRecord in items: 
      ret.append(self.normalizeItem(itemRecord))

    print("sort: " + args["sort"])
    if args['sort'] in pythonSorts:
        print("Doing a python sort!")
        ret.sort(key = lambda x: x[args['sort']], reverse = True)

    return self.outputJson(ret);

  @cherrypy.expose
  def changePassword(self, *path, **args):
    self.wrapper.changePassword(self.getUsername(), args['hashedPassword']);

    return self.outputJson(JSON_OK)

  @cherrypy.expose
  def deleteTask(self, *path, **args):
    self.wrapper.deleteTask(int(args['id']))

    return self.outputJson(JSON_OK);

  @cherrypy.expose
  def deleteTag(self, *path, **args):
    self.wrapper.deleteTag(int(args['id']))

    return self.outputJson(JSON_OK)

  @cherrypy.expose
  def renameItem(self, *path, **args):
    self.wrapper.setTaskContent(int(args['id']), args['content']);
    return self.outputJson(JSON_OK);

  @cherrypy.expose
  def deleteList(self, *path, **args):
    self.wrapper.deleteList(int(args['id']));

    return self.outputJson(JSON_OK);

  def outputJsonError(self, code, msg, uniqueType = ""):
    cherrypy.response.status = code;

    return self.outputJson({
      "type": "Error",
      "uniqueType": uniqueType,
      "message": msg
    })

  @cherrypy.expose
  def register(self, *path, **args):
    try:
      user, password = api.wrapper.getUser(args['username']);

      if user != None:
        return self.outputJsonError(403, "User already exists", uniqueType = "username-already-exists");

      if len(args['username']) < 3:
        return self.outputJsonError(403, "Username must be at least 3 characters long.", uniqueType = "username-too-short")

      if len(args['password']) < 6:
        return self.outputJsonError(403, "Password must be at least 6 character long.", uniqueType = "password-too-short")

      api.wrapper.register(args['username'], args['password'], args['email'])

      return self.outputJson(JSON_OK);
    except Exception as e:
      return self.outputJsonError(403, str(e))

  def randomWallpaper(self):
    wallpapers = []

    try:
      wallpapers = []
      
      for wallpaper in os.listdir(args.wallpaperdir):
        if wallpaper.endswith(".png") or wallpaper.endswith(".jpg") or wallpaper.endswith(".webp"):
          wallpapers.append(wallpaper)
    except Exception as e:
      print(e)

    if len(wallpapers) == 0:
      return None
    else:
      return random.choice(wallpapers);

  @cherrypy.expose
  def init(self, *path, **args):
    username = None

    if "username" not in cherrypy.session:
      cherrypy.session['username'] = None
      cherrypy.session.save();
    else:
      username = cherrypy.session['username']

    return self.outputJson({
      "wallpaper": self.randomWallpaper(),
      "id": cherrypy.session.id,
      "username": username
    });

  def isLoggedIn(self):
    if "username" in cherrypy.session:
      if cherrypy.session['username'] != None:
        return True

    return False

  def checkLoggedIn(self):
    if not self.isLoggedIn():
      raise cherrypy.HTTPError(403, "Login required.")
  
  def getUsername(self):
    self.checkLoggedIn();

    return cherrypy.session['username']

  @cherrypy.expose
  def updateTag(self, *path, **args):
    updatedTag = self.wrapper.updateTag(int(args['id']), args['newTitle'], args['shortTitle'], args['backgroundColor'], args["textualValue"], args["numericValue"], int(args['tagValueId']))

    return self.outputJson(updatedTag)


  @cherrypy.expose
  def authenticate(self, *path, **args):
    HttpQueryArgChecker(args).requireArg('username');

    user, password = api.wrapper.getUser(args['username']);

    if user == None:
      return self.outputJsonError(403, "User not found: " + args['username'], uniqueType = "user-not-found")

    if args['password'] != password:
      return self.outputJsonError(403, "Password is incorrect.", uniqueType = "user-wrong-password")

    cherrypy.session.regenerate();
    cherrypy.session['username'] = user['username'];
    cherrypy.session.save();

    return self.outputJson({
      "username": user['username'],
      "id": cherrypy.session.id,
    });

  @cherrypy.expose
  def logout(self, *path, **args):
    cherrypy.session['username'] = None
    cherrypy.session.regenerate();

    return self.outputJson({"message": "Logged out!"})
    
def CORS():
  print("Registering CORS handler")
  cherrypy.response.headers['Access-Control-Allow-Origin'] = args.corsDomain
  cherrypy.response.headers['Access-Control-Allow-Credentials'] = "true"

def http_error_handler(status, message, traceback, version):
  return json.dumps({
    "httpStatus": status, 
    "type": "httpError",
    "message": message
  });

  
def error_handler():
  cherrypy.response.status = 500;
  cherrypy.response.headers['Content-Type'] = "text/plain"

  exceptionInfo = exc_info()
  excType = exceptionInfo[0]
  exception = exc_info()[1]

  # Clients get a simple version.
  return "\nUnhandled exception.\n" + "Message: " + str(exception) + "\n" + "Type: " + str(excType.__name__)

  # Logs get a full version
  print(exceptionInfo)

api = Api();

cherrypy.config.update({
  'server.socket_host': '0.0.0.0',
  'server.socket_port': args.port,
  'tools.sessions.on': True,
  'tools.sessions.locking': 'early',
#  'tools.sessions.storage_class': cherrypy.lib.sessions.FileSession,
#  'tools.sessions.storage_path': '/tmp/wt-sessions',
  'tools.sessions.timeout': 20160,
  'tools.CORS.on': True,
  'request.error_response': error_handler,
  'request.error_page': {'default':  http_error_handler}
});

cherrypy.tools.CORS = cherrypy.Tool('before_handler', CORS);

if args.background:
  cherrypy.process.plugins.Daemonizer(cherrypy.engine).subscribe()

cherrypy.quickstart(api)
