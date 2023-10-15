## ShotGoods

### Server Parameter Sheet
| |game_biz|region|
|--|--|--|
|Genshin Impact|hk4e_cn|cn_gf01|
|Honkai Impact 3rd|bh3_cn|pc01|
|Houkai Gakuen 2|bh2_cn|
|Tears of Themis|nxx_cn|
|Honkai: Star Rail|hkrpg_cn|prod_gf_cn|

### Cookie
Necessary cookie: `account_id` and `cookie_token`

## Usage
> Linux & virtual exchange
```shell
./ShotGoods -type "virtual" -id "2023022412691" -uid "100000000" -region "cn_gf01" -game_biz "hk4e_cn" -target "1678882996" -config "config.json"
```

> Windows & real exchange
```bat
.\ShotGoods.exe -type "real" -id "2023022412691" -target "1678882996" -config "config.json"
```