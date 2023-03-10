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

## Usage
```go
func main() {
	ReadConfig()
	// 奖品详情: https://github.com/jellyqwq/ShotGoods/blob/main/goods.csv
    
	// 实物兑换
	// good := NewRealGood("2023022311902", 1, config.AddressId)
	// 游戏内兑换 (原神为例)
	good := NewVirtualGood("2023022412691", 1, "Yuanshen uid", "cn_gf01", "hk4e_cn")

	// good.Worker(getTime("19:00:00"))
    // Goods的 next_time
	good.Worker(parseUnix("1678878000"))
}
```