#!/usr/bin/python3

import __init__
import argparse 
import wrapper
import hashlib
import commonArgumentParser

parser = commonArgumentParser.getNew()
parser.add_argument('--username', '-u', required = True)
parser.add_argument("--password", required = True)
args, unknown = parser.parse_known_args();

m = hashlib.sha1()
m.update(args.password.encode())

api = wrapper.fromArgs(args);
res = api.changePassword(args.username, m.hexdigest());

print("Rows changed: " + str(len(res.values())))
