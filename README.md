## ShotGoods

### 表单填写
| |uid|game_biz|region|
|--|--|--|--|
|原神|略|hk4e_cn|cn_gf01|
|崩坏3|略|bh3_cn|pc01|
|崩坏2|略|bh2_cn||
|未定事件簿|略|nxx_cn

region得看自己服务器

### cookie
必要cookie是 `account_id` 和 `cookie_token`

如果要获取对应的游戏服务器还需要 `stoken` 和 `stuid` (暂未实现)

## Usage
> Linux & virtual exchange
```shell
./ShotGoods -type "virtual" -id "2023022412691" -uid "100000000" -target "1678882996" -conf "config.json"
```

> Windows & real exchange
```shell
ShotGoods.exe -type "real" -id "2023022412691" -target "1678882996" -conf "config.json"
```