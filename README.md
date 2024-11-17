# Abashiri-CLI
Abashiri-CLI automates subdomain enumeration and URL enumeration for each subdomain, managing the results in a database.

```
$  ./abashiri-cli 
. . 
し  < ABASHIRI-CLI!!!
 ▽
```

## Usage

### Scan
```
パッシブスキャン dns bruteforceを行わない
$ ./abashili-cli scan -d example.com

アクティブスキャン dns bruteforceも行う
$ ./abashili-cli scan -d example.com -m active
```

### Show

```
DBに登録されたルートドメインの表示
$ ./abashili-cli show domain --root

指定したドメインに紐づくサブドメインの表示
$ ./abashili-cli show domain -d example.com

指定したドメインに関連するサブドメインのURLを表示
$ ./abashili-cli show url -d example.com
```


### Delete
```
ルートドメインをSQLから削除
$ ./abashili-cli delete -d example.com
```


## 現在の依存関係　
以下のツールが実行可能な環境であること
- [subfinder](https://github.com/projectdiscovery/subfinder)
- [dnsx](https://github.com/projectdiscovery/dnsx)

---

## TODO

- [ ] サブドメイン列挙
現状subfinderに依存しているので棚卸しする

- [ ] URL列挙
  - [x] Wayback Machine
  - [x] Common Crawl
  - [ ] OTX AlienVault
  - [ ] urlscan.io
  - [ ] Search Engines(Google, DuckDuckGo, Bing)

- [ ] 操作関連
  - [ ] 全ドメインの表示
  - [ ] scan時にカンマ区切りでドメインを渡せるようにする

- [ ] その他
  - [ ] core/discovery/url_scan.goのテスト
  - [ ] ドメインの親子関係の整合性確認ロジック
  - [ ] 管理IPアドレス収集のためにDB設計


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



