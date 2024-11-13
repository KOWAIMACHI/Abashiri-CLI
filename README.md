# Abashiri-CLI
## 🚧work in progress🚧

AmassやSubfinderなどの既存のCLIツールを実行し、収集したサブドメインやWebサービスの情報をSqlite3データベースで管理します。
今後ちょっとずつ自前実装に変えていきたい実装していきたい。

## Usage
```
$  ./abashiri-cli 
. . 
し  < ABASHIRI-CLI!!!
 ▽

Usage:
  abashiri-cli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  scan        Scan a domain for subdomains using passive or active methods
  show        A brief description of your command

Flags:
  -h, --help      help for abashiri-cli
  -v, --verbose   Enable verbose output

Use "abashiri-cli [command] --help" for more information about a command
```
### Scan

```
$ ./abashili-cli scan -d example.com
$ ./abashili-cli scan -d example.com -m active -v
```

### Show
```
$ ./abashili-cli show links -d example.com
```

## メモ

## 全体設計

```
[discovery] -> [save] -> [filter] -> [show]
```

```
├─ cmd
│   └─ subdomain
├─ core
│   └─ discovery             // core
│       └─ subdomain_scan.go
│       └─ xxxx_scan.go
├─ helpers                   // 補助機能やユーティリティ関数
├─ storage                   // データ保存・管理の機能
│   └─ subdomain_storage.go
│   └─ xxxx_storage.go
├─ lib                          // いろんな外部ツールに依存する予定なので、この辺にコンパイル済みバイナリを置いて依存関係減らしたい。(OR Dockerにする)
├─ wordlists
│   └─ dns
│       └─ subdomains-top1million-2000.txt
```


## DB設計(今後)

```
- domains (基本的に最初の入力値の予定)
  - id (PK)
  - name
  - corp_id (FK)
  - created_at
  - updated_at

- subdomains
  - id (PK)
  - parent_id (PK)
  - root_id (PK)
  - name
  - tools_detected  // どのツールによって検出したか
  - created_at
  - updated_at

- links
  - id
  - url 
  - domain_id (対応するdomain)
  - subdomain_id (対応するsubdomain) //これがNULLだとroot domainに紐付いたリンクということになる 
  - created_at
  - updated_at
```


---

## Prerequirement
以下のツールが実行可能な環境であること
- [amass](https://github.com/owasp-amass/amass)
- [subfinder](https://github.com/projectdiscovery/subfinder)
- [dnsx](https://github.com/projectdiscovery/dnsx)


