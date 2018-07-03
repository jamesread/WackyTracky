#!/usr/bin/python3

import __init__
import argparse 
import wrapper

parser = argparse.ArgumentParser();
parser.add_argument('--userid', '-u', type = int, required = True)
args = parser.parse_args();

api = wrapper.Wrapper();

api.deleteUser(args.userid);
