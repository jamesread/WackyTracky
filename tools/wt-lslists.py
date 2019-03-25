#!/usr/bin/python3

import __init__
import argparse 
import wrapper
from prettytable import PrettyTable

import commonArgumentParser

parser = commonArgumentParser.getNew()
parser.add_argument('--username', '-u', required = True)
args = parser.parse_args();

api = wrapper.fromArgs(args);

table = PrettyTable(['ID', 'Title'])

for ret in api.getLists(args.username):
	tasklist = ret[0]

	table.add_row([tasklist.id, tasklist['title']])

print(table)


