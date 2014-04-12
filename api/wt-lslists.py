#!/usr/bin/python

import argparse 
import wrapper
from prettytable import PrettyTable

parser = argparse.ArgumentParser();
parser.add_argument('--username', '-u', required = True)
args = parser.parse_args();

api = wrapper.Wrapper();

table = PrettyTable(['ID', 'Title'])

for ret in api.getLists(args.username):
	tasklist = ret[0]

	table.add_row([tasklist.id, tasklist['title']])

print table


