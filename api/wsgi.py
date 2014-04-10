#!/usr/bin/python

import cherrypy
import atexit

def CORS():
	cherrypy.response.headers['Access-Control-Allow-Origin'] = "http://www2.teratan.net/"
	cherrypy.response.headers['Access-Control-Allow-Credentials'] = "true"

cherrypy.tools.CORS = cherrypy.Tool('before_handler', CORS);
cherrypy.config.update({
	'environment': 'embedded',
	"log.error_file": "/tmp/site.log"
});

if cherrypy.engine.state == 0:
	cherrypy.engine.start(blocking = False)
	atexit.register(cherrypy.engine.stop)

class Root:
	def __init__(self):
		print "init"

	@cherrypy.expose
	def index(self, *args, **kwargs):
		print "index"
		return "hi";

	@cherrypy.expose
	def default(self):
		print "default1"
		return "defaulggt"

	index.exposed = True
	default.exposed = True

def application(environ, start_response):
	print "app()"

	cherrypy.tree.mount(Root())
	return cherrypy.tree(environ, start_response);
