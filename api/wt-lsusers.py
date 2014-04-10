#!/usr/bin/python

import wrapper
from prettytable import PrettyTable

api = wrapper.Wrapper();

table = PrettyTable(['ID', 'Username', 'Email'])

for ret in api.getUsers():
	user = ret[0]

	table.add_row([user.id, user['username'], user['email']])

print table
