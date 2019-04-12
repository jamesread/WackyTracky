#!/usr/bin/env python3

import __init__
import wrapper
import commonArgumentParser

parser = commonArgumentParser.getNew()
args, _ = parser.parse_known_args();

print(args.dbUser)

api = wrapper.fromArgs(args)

cql = "MATCH (l:List)-[:scheduled]->(tv:TagValue) OPTIONAL MATCH (l)-->(i:Item) CREATE UNIQUE (i)-[:tagged]->(tv) ";
res = api.session.run(cql);

for row in res:
    print("row")
    print(row)


