# Abashiri-CLI
## 🚧work in progress🚧

AmassやSubfinderなどの既存のCLIツールを実行し、収集したサブドメインやWebサービスの情報をSqlite3データベースで管理します。

## Prerequirement
以下のツールが実行可能な環境であること
- [amass](https://github.com/owasp-amass/amass)
- [subfinder](https://github.com/projectdiscovery/subfinder)
- [dnsx](https://github.com/projectdiscovery/dnsx)

## Usage

### subdomain

- **subdomain scan**
```
# passive
abashiri subdomain scan --domain loom.com --mode passive -v
# active (dns bruteforce)
abashiri subdomain scan --domain loom.com --mode active -v
```

- **get data**
```
abashiri subdomain get --domain example.com
```

