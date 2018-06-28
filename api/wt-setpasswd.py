#!/usr/bin/python3

import argparse 
import wrapper

parser = argparse.ArgumentParser();
parser.add_argument('--userid', '-u', required = True)
parser.add_argument("--password", required = True)
args = parser.parse_args();

api = wrapper.Wrapper();
api.changePassword(args.userid, args.password);
