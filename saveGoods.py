# https://api-takumi.miyoushe.com/mall/v1/web/goods/list?app_id=1&point_sn=myb&page_size=20&page=1&game=hk4e
import requests
import time
import json
import platform
from proxy import ProxyServer

proxies = None
if platform.system() =='Windows':
    ps = ProxyServer()
    if ps.is_open_proxy_form_Win():
        ip, port = ps.get_server_form_Win()
        proxies = {
            'http': f'http://{ip}:{port}',
            'https': f'http://{ip}:{port}'
        }

# 先获取要抢的奖励
BaseURL = 'https://api-takumi.miyoushe.com/mall/v1/web/goods'
GetGoodsList = BaseURL + '/list'
PostExchangeGoods = BaseURL + '/exchange'

class Client:
    def __init__(self) -> None:
        with open("config.json", "r", encoding="UTF-8") as f:
            self.config = json.loads(f.read())
        self.headers = {
            "Accept": "application/json, text/plain, */*",
            "Accept-Language": "en-US,en-GB;q=0.9,en;q=0.8",
            "Accept-Encoding": "gzip, deflate, br",
            "Content-Type": "application/json;charset=utf-8",
            "x-rpc-device_model": self.config["x-rpc-device_model"],
            "x-rpc-device_fp": self.config["x-rpc-device_fp"],
            "x-rpc-client_type": self.config["x-rpc-client_type"],
            "x-rpc-device_id": self.config["x-rpc-device_id"],
            "x-rpc-channel": self.config["x-rpc-channel"],
            "x-rpc-app_version": self.config["x-rpc-app_version"],
            "x-rpc-device_name": self.config["x-rpc-device_name"],
            "x-rpc-sys_version": self.config["x-rpc-sys_version"],
            "Origin": "https://webstatic.miyoushe.com",
            "Referer": "https://webstatic.miyoushe.com/",
            "User-Agent": self.config["User-Agent"],
            "Cookie": self.config["Cookie"],
        }
        self.GoodsData = []
        self.GoodsDataGBK = []

    def SaveGoodsList(self, pageNumber=1):
        url =  GetGoodsList + f'?app_id=1&point_sn=myb&page_size=20&page={pageNumber}&game='
        response = requests.get(url, headers=self.headers, proxies=proxies)
        if response.status_code != 200:
            print('requests error')
            return
        jsonContent = response.json()
        if jsonContent['retcode'] != 0:
            print('retcode error')
            return
        
        if pageNumber == 1:
            title = list(jsonContent['data']['list'][0].keys())
            self.GoodsData.append(title)
            self.GoodsDataGBK.append(title)

        for i in jsonContent['data']['list']:
            tempList = []
            tempListGBK = []
            for c, v in enumerate(i.values()):
                if c == 1:
                    tempList.append(str(v))
                    tempListGBK.append(f"'{v}")
                else:
                    tempList.append(str(v))
                    tempListGBK.append(str(v))
            self.GoodsData.append(tempList)
            self.GoodsDataGBK.append(tempListGBK)

        total = jsonContent['data']['total']
        if pageNumber * 20 < total:
            self.SaveGoodsList(pageNumber+1)
            time.sleep(1)
        else:
            with open('goods.csv', 'w+', encoding='utf-8') as f:
                for row in self.GoodsData:
                    f.write(','.join(row))
                    f.write('\n')
            with open('goods_gbk.csv', 'w+', encoding='gbk') as f:
                for row in self.GoodsDataGBK:
                    f.write(','.join(row))
                    f.write('\n')

    def getRoleGame(self):
        # 需要ds, 没签不管了, 这个用不了 [需要stuid和stoken]
        headers = self.headers.copy()
        headers['Refer'] = 'https://app.mihoyo.com'
        headers['Cookie'] = ''
        response = requests.get('https://api-takumi-record.mihoyo.com/game_record/app/card/wapi/getGameRecordCard?uid=288798795', headers=self.headers, proxies=proxies)
        print(response.text)

if __name__ == '__main__':
    c = Client()
    c.SaveGoodsList()
    # c.startShot()
