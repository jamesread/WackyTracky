#!/usr/bin/python3

import __init__
import argparse
import wrapper

import commonArgumentParser

parser = commonArgumentParser.getNew()
parser.add_argument("--username", required = True)
args = parser.parse_args()

api = wrapper.fromArgs(args)

api.createUser(args.username)
