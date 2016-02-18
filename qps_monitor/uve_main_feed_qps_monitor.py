#!/usr/local/sina_mobile/python2.7/bin/python

import time
import json
import redis
import urllib2
import logging

FRONT_REDIS_MASTER_HOST = '10.13.40.50'
FRONT_REDIS_MASTER_PORT = 6379
PERCENTAGE_OF_DECREASE = 0.4
GRAPHITE_URL = 'http://10.77.96.122:8085/render/'

class Monitor:

    def __init__(self, logger = None, interval = 5):
        self.logger = logger or logging.getLogger(__name__)
        self.interval = interval
        redis_pool = redis.ConnectionPool(host = FRONT_REDIS_MASTER_HOST, port = FRONT_REDIS_MASTER_PORT, db = 0)
        self.redis = redis.Redis(connection_pool = redis_pool)
        self.logger.info('UVE monitor is started')
    
    def run(self):
        while True:
            try:
                self.stats()
            except Exception, e:
                self.logger.error('Monitor stats failed', exc_info = True)
            time.sleep(self.interval)

    def stats(self):
        url = GRAPHITE_URL + '?target=uve.access.qps.uve-service-main_feed&format=json&from=-300s'
        response = urllib2.urlopen(url, timeout=3)
        json_string = response.read()
        parsed_json = json.loads(json_string)

        mainfeed = 0
        data = parsed_json[0]['datapoints']
        for _data in data:
            if _data[0] != None:
                mainfeed = int(_data[0])
        self.save_data(mainfeed)

    def save_data(self, mainfeed):
        self.logger.info('main_feed qps:' + str(mainfeed))
        key = 'uve_main_feed_qps2'
        _last = self.redis.get(key)
        _p = 0
        if _last:
            _last = int(_last)
            _p = (_last - mainfeed) / _last

        if _p < PERCENTAGE_OF_DECREASE:
            self.redis.set(key, mainfeed)
        else:
            self.logger.warn('Did NOT update mainfeed QPS')

if __name__ == '__main__':
    logger = logging.getLogger(__name__)
    logger.setLevel(logging.INFO)

    handler = logging.FileHandler('uve_main_feed_qps_monitor.log')
    handler.setLevel(logging.INFO)

    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)

    monitor = Monitor(logger, 5)
    monitor.run()
