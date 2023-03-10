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

### 实物
```
Goods{
    App_id: 1,
    Point_sn: "myb",
    Exchange_num: 1,
    Goods_id: "xxx",
    Address_id: xxx,
}
```

### 虚拟物品
```
Goods{
    App_id: 1,
    Point_sn: "myb",
    Exchange_num: 1,
    Goods_id: "xxx",
    Uid: "xxx",
    Region: "xxx",
    Game_biz: "xxx",
    Address_id: 0,
}
```