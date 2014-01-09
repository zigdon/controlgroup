#!/usr/bin/python
"""
Python MapReduce - A very basic way to run mass operations on text.

Example usages:

  Basic map operation -

    $ echo -e "1\n2\n3" | ./pythonmr.py --map="int(item)*2"
    2
    4
    6

  Basic reduce operation (with initialized accumulator) -

    $ echo -e "1\n2\n3\n4" | ./pythonmr.py --reduce="int(item)+accum" --accum="0"
    10

  Without --skip, this would result in an error (due to the blank line where 3 was),
  but with it, the blank line is ignored -

    $ echo -e "1\n2\n\n4" | ./pythonmr.py --skip --reduce="int(item)+accum" --accum="0"
    10

  Combined map and reduce -

    $ echo -e "1\n2\n3\n4" | ./pythonmr.py --map="int(item)" --reduce="item+accum"
    10
"""
import argparse
import collections
import itertools
import sys

def map_python(expr, iterable):
    def eval_expr(item):
        return eval(expr, {"item": item})
    return itertools.imap(eval_expr, iterable)

def reduce_python(expr, iterable, initial):
    def eval_expr(accum, item):
        return eval(expr, {"item": item, "accum": accum})
    if initial is not None:
        return reduce(eval_expr, iterable, eval(initial))
    else:
        return reduce(eval_expr, iterable)

def main():
    parser = argparse.ArgumentParser(description="Process text by running Python code as map and reduce steps.")
    parser.add_argument("-m", "--map", metavar="EXPR",
        help="A python expression to be mapped onto the input lines (use 'item' variable).")
    parser.add_argument("-s", "--skip", action="store_true",
        help="Omit falsey (None, False, empty string, etc) values after map step.")
    parser.add_argument("-r", "--reduce", metavar="EXPR",
        help="A python expression to be reduced onto the input lines (use 'accum' and 'item' variables).")
    parser.add_argument("-a", "--accum", metavar="EXPR",
        help="A python expression with which to initialize the accumulator for reduces.")
    parser.add_argument("-i", "--in", metavar="FILEPATH",
        help="A path to a file to use as input, instead of stdin.")
    parser.add_argument("-o", "--out", metavar="FILEPATH",
        help="A path to a file to use as output, instead of stdout.")

    args = vars(parser.parse_args())

    # Default to stdin as the data source
    if args["in"]:
        source = open(args["in"])
    else:
        source = sys.stdin

    # First, strip the newlines off each line of text
    result = itertools.imap(lambda line: line.rstrip("\n"), source)

    # Then, run a map step, if specified
    if args["map"]:
        result = map_python(args["map"], result)

    if args["skip"]:
        result = (item for item in result if item)

    # Then a reduce step, if specified
    if args["reduce"]:
        result = reduce_python(args["reduce"], result, args["accum"])

    # Finally, output the result
    if args["out"]:
        output = open(args["out"], "w")
    else:
        output = sys.stdout

    if isinstance(result, collections.Iterable):
        for item in result:
            output.write("%s\n" % item)
    else:
        output.write("%s\n" % result)


if __name__ == "__main__":
    main()

# vim: set ts=4 sts=4 sw=4 et:
