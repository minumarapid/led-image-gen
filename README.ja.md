# led-image-gen
発光エフェクト付きLED風画像を生成するGolang製の画像処理ツールです。  
現状、WebAPI・CLIで使用できます。

## 特徴
- 発光エフェクト付きLED風画像を生成
- WebAPIとCLIで利用可能
- GoRoutineによる並列処理を活用した高速な画像処理

## 必要条件
- Go 1.25以降

## Configuration
| パラメータ名          | 説明                           | 型          | 必須? | デフォルト値                      |
|-----------------|------------------------------|------------|-----|-----------------------------|
| `Border`        | 画像の外枠の幅                      | int        | いいえ | 10                          |
| `LEDSize`       | LEDのサイズ                      | int        | いいえ | 4                           |
| `LEDGap`        | LED同士の間隔                     | int        | いいえ | 2                           |
| `LEDGamma`      | LEDのガンマ補正値                   | float64    | いいえ | 1.0                         |
| `LEDExposure`   | LEDの露出補正値                    | float64    | いいえ | 1.0                         |
| `LEDShape`      | LEDの形状（true: 円形、false: 四角形）  | bool       | いいえ | false                       |
| `MaxWorkers`    | 並列処理に使用する最大ワーカー数             | int        | いいえ | 4                           |
| `EnableGlow`    | 発光エフェクトの有効化                  | bool       | いいえ | true                        |
| `GlowRange`     | 発光エフェクトの範囲                   | float64    | いいえ | 1.0                         |
| `GlowStrength`  | 発光エフェクトの強さ                   | float64    | いいえ | 1.75                        |
| `GlowGamma`     | 発光エフェクトのガンマ補正値               | float64    | いいえ | 1.0                         |
| `GlowExposure`  | 発光エフェクトの露出補正値                | float64    | いいえ | 1.0                         |
| `OffLightColor` | LEDがオフ(RGBA{0,0,0,255})のときの色 | color.RGBA | いいえ | color.RGBA{40, 40, 40, 255} |

## API
### ベースURL
Public (Vercel Serverless Functions)
```
https://api.led.o38.me/
```

Local
```
http://localhost:8080/
```

### エンドポイント
- `POST /api/` - 画像生成エンドポイント(form-dataで画像ファイルと設定を受け取ります)

## ライセンス
このプロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。