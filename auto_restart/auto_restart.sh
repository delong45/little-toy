#!/bin/bash

process="newstats"
pid_num=2
error_num=0
max_error_num=3

is_access_running=true
is_subreq_running=true

start_access_cmd="nohup python /usr/home/delong1/newstats.py -f /data0/nginx/logs/mobiletrends.mobile.sina.cn_access.log &"
start_subreq_cmd="nohup python /usr/home/delong1/newstats.py -f /data0/nginx/logs/mobiletrends.mobile.sina.cn_subreq.log -p 3333 -c subreq &"

function restart() {
    local category=$1
    if [ $category = "access" ]; then
        eval $start_access_cmd
        is_access_running=true
    elif [ $category = "subreq" ]; then
        eval $start_subreq_cmd
        is_subreq_running=true
    fi
}

function check() {
    local access_stats_pid=$(ps -ef | grep -v "grep" | grep $process | grep "access" | awk '{print $2}')
    local access_pid_num=$(ps -ef | grep -v "grep" | grep $process | grep "access" | awk '{print $2}' | wc -l)
    if [ $access_pid_num -ne 1 ]; then
        for i in $access_stats_pid; do
            sudo kill -9 $i
        done
        is_access_running=false
        restart "access"
    fi

    local subreq_stats_pid=$(ps -ef | grep -v "grep" | grep $process | grep "subreq" | awk '{print $2}')
    local subreq_pid_num=$(ps -ef | grep -v "grep" | grep $process | grep "subreq" | awk '{print $2}' | wc -l)
    if [ $subreq_pid_num -ne 1 ]; then
        for i in $subreq_stats_pid; do
            sudo kill -9 $i
        done
        is_subreq_running=false
        restart "subreq"
    fi
}

#main
check
