#!/usr/bin/python3

import __init__
import wrapper
from prettytable import PrettyTable

import commonArgumentParser

parser = commonArgumentParser.getNew()
args, _ = parser.parse_known_args()
api = wrapper.fromArgs(args)

table = PrettyTable(['ID', 'Username', 'Email', 'Lists', 'Items'])

for ret in api.getUsers():
	user = ret[0]

	table.add_row([user.id, user['username'], user['email'], ret[1], ret[2]])

print(table)
