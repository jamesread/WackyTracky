#!/usr/bin/python3

import __init__
import argparse
import wrapper

parser = argparse.ArgumentParser()
parser.add_argument("--username", required = True)
args = parser.parse_args()

api = wrapper.Wrapper()

api.createUser(args.username)
