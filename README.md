# Abashiri-CLI
Abashiri-CLI automates subdomain enumeration and URL enumeration for each subdomain, managing the results in a database.

```
$  ./abashiri-cli 
. . 
し  < ABASHIRI-CLI!!!
 ▽
```

## Usage
```bash
# サブドメインの列挙+URLの列挙
abashiri all -d example.com
# サブドメインの列挙(passive)
abashiri subdomain -d example.com
# サブドメインの列挙(active)
abashiri subdomain -d example.com -m active
# 登録されているサブドメインのURL列挙(passive)
abashiri url -d example.com

# HTMLエクスポート
abashiri export -d example.com

# 登録されているドメインの表示
abashiri show domain -r
# サブドメインの表示
abashiri show domain -d example.com
# 各サブドメインのURL表示
abashiri show url -d example.com
```


## 現在の依存関係　
以下のツールが実行可能な環境であること
- [subfinder](https://github.com/projectdiscovery/subfinder)
- [dnsx](https://github.com/projectdiscovery/dnsx)

