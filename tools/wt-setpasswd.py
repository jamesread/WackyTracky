#!/usr/bin/python3

import __init__
import argparse 
import wrapper

import commonArgumentParser

parser = commonArgumentParser.getNew()
parser.add_argument('--username', '-u', required = True)
parser.add_argument("--password", required = True)
args = parser.parse_args();

api = wrapper.fromArgs(args);
res = api.changePassword(args.username, args.password);

print("Rows changed: " + str(len(res.values())))
