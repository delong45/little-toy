import json

f = open('tmp.txt', 'r')
log = f.read()
log = log.strip('\n')

level_1_FS = '\x1c'

level_2_SH = '\x01'
level_2_GS = '\x1d'

level_3_RS = '\x1e'
level_3_US = '\x1f'

log_dict = {}
log_list = log.split(level_1_FS)
for item in log_list:
    item = item.split(':', 1)
    if item[0] == 'tmeta_l2':
        str_l2 = item[1]
    log_dict[item[0]] = item[1]

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
log_json = json.dumps(log_dict, sort_keys=True, indent=4)
print log_json
