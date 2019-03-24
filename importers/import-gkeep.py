#!/usr/bin/python3

import __init__
import wrapper
import gkeepapi

import configargparse

parser = configargparse.ArgumentParser(default_config_files=["/etc/wacky-tracky/import-gkeep.cfg"])
parser.add_argument("--username", required = True)
parser.add_argument("--password", required = True)
args = parser.parse_args()

print(args.username);

keep = gkeepapi.Keep()

success = keep.login(args.username, args.password)

note = gkeep.createNode("Todo", "Eat breakfast")
keep.sync()
