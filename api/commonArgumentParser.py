import configargparse

def getNew():
    parser = configargparse.ArgParser(default_config_files=["/etc/wacky-tracky/server.cfg", "~/.wacky-tracky/server.cfg"]);
    parser.add_argument("--dbDriver", env_var = "WT_DB_DRIVER", default = "mysql")
    parser.add_argument("--dbUser", env_var = "WT_DB_USER")
    parser.add_argument("--dbPassword", env_var = "WT_DB_PASS")
    parser.add_argument("--dbServer", env_var = "WT_DB_USER", default = "localhost")

    return parser;

def getParsed():
    args, _ = getNew().parse_known_args()

    return args
