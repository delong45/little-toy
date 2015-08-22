#!/usr/local/sina_mobile/python2.7/bin/python

# @author Li Delong<delong1@staff.weibo.com>
# @copyright (C) Weibo.com 2015 

import json
import os
import sys
import fcntl

def stats_to_json(log):
    level_1_FS = '\x1c'

    level_2_SH = '\x01'
    level_2_GS = '\x1d'

    level_3_RS = '\x1e'
    level_3_US = '\x1f'

    if 'tmeta_l2' in log:
        has_level_2 = True
    else:
        has_level_2 = False

    log_dict = {}
    log_list = log.split(level_1_FS)
    for item in log_list:
        item = item.split(':', 1)
        if item[0] == 'tmeta_l2':
            str_l2 = item[1]
        log_dict[item[0]] = item[1]

    if not has_level_2:
        return log_dict

    #process tmeta_l2
    str_l2 = str_l2.split(level_2_SH)
    for index in range(len(str_l2)):
        str_l2[index] = str_l2[index].split(level_2_GS)
        dict = {}
        for item in str_l2[index]:
            item = item.split(':', 1)
            key = item[0]
            value = item[1]

            #process tmeta_l3
            if key == 'tmeta_l3':
                value = value.split(level_3_RS)
                for i in range(len(value)):
                    value[i] = value[i].split(level_3_US)
                    d = {}
                    for it in value[i]:
                        it = it.split(':')
                        d[it[0]] = it[1]
                    value[i] = d
            #end

            dict[key] = value
        str_l2[index] = dict
    #end

    log_dict['tmeta_l2'] = str_l2
    return log_dict

def set_stdin_nonblock():
    fd = sys.stdin.fileno()
    fl = fcntl.fcntl(fd, fcntl.F_GETFL)
    fcntl.fcntl(fd, fcntl.F_SETFL, fl | os.O_NONBLOCK)

if __name__ == '__main__':
    try:
        set_stdin_nonblock()
        lines = sys.stdin.readlines()
    except:
        if len(sys.argv) < 2:
            print "Usage:", sys.argv[0], "[filename]"
            sys.exit(1)

        if os.path.isfile(sys.argv[1]):
            lines = open(sys.argv[1], 'r')
        else:
            print "Error:", sys.argv[0], "doesn't exsit file", sys.argv[1]
            sys.exit(2)
        
    for line in lines:
        line = line.strip('\n')
        try:
            log_dict = stats_to_json(line)
            log_json = json.dumps(log_dict, sort_keys = True, indent = 4)
            print log_json + '\n'
        except:
            print "Error:", sys.argv[0], "cann't convert to json"
