#!/usr/bin/env python
import sys
import os

def process(filename, f):
    string = os.path.abspath(filename)
    end_pos = string.rfind('.')
    if end_pos == -1:
        output_name = string + '_rookie'
    else:
        output_name = string[:end_pos] + '_rookie' + string[end_pos:]
    output = open(output_name, 'w')

    for line in f:
        if line != '\n':
            line = line.replace('\n', ' ')
        output.write(line)
    output.close() 
    print "Formating into", output_name

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print "Usage:", sys.argv[0], "[file]"
        sys.exit(1)

    if os.path.isfile(sys.argv[1]):
        filename = sys.argv[1]
        f = open(filename, 'r')
    else:
        print "Error:", sys.argv[0], "doesn't exsit file", sys.argv[1]
        sys.exit(2)

    process(filename, f)
    f.close()
