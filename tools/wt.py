#!/usr/bin/env python3

import click

@click.command()
@click.option("--name", prompt = "Your name")
def hello(name):
    print("hello", name)

if __name__ == '__main__':
    hello()
