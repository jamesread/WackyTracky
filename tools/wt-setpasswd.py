#!/usr/bin/python3

import __init__
import argparse 
import wrapper

parser = argparse.ArgumentParser();
parser.add_argument('--username', '-u', required = True)
parser.add_argument("--password", required = True)
args = parser.parse_args();

api = wrapper.Wrapper();
res = api.changePassword(args.username, args.password);

print("Rows changed: " + str(len(res.values())))
