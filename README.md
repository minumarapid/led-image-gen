# led-image-gen
A Golang image processing tool that generates LED-style images with glow effects.  
Currently available as WebAPI and CLI.

## Features
- Generates LED-style images with glow effects
- Available as WebAPI and CLI
- Fast image processing using GoRoutine parallel processing

## Requirements
- Go 1.25 or later

## Configuration
| Parameter Name  | Description                                       | Type       | Required? | Default Value               |
|-----------------|---------------------------------------------------|------------|-----------|-----------------------------|
| `Border`        | Width of the image border                         | int        | No        | 10                          |
| `LEDSize`       | Size of the LED                                   | int        | No        | 4                           |
| `LEDGap`        | Gap between LEDs                                  | int        | No        | 2                           |
| `LEDGamma`      | Gamma correction value for LED                    | float64    | No        | 1.0                         |
| `LEDExposure`   | Exposure correction value for LED                 | float64    | No        | 1.0                         |
| `LEDShape`      | Shape of LED (true: circle, false: square)        | bool       | No        | false                       |
| `MaxWorkers`    | Maximum number of workers for parallel processing | int        | No        | 4                           |
| `EnableGlow`    | Enable glow effect                                | bool       | No        | true                        |
| `GlowRange`     | Range of glow effect                              | float64    | No        | 1.0                         |
| `GlowStrength`  | Strength of glow effect                           | float64    | No        | 1.75                        |
| `GlowGamma`     | Gamma correction value for glow                   | float64    | No        | 1.0                         |
| `GlowExposure`  | Exposure correction value for glow                | float64    | No        | 1.0                         |
| `OffLightColor` | Color when LED is off (RGBA{0,0,0,255})           | color.RGBA | No        | color.RGBA{40, 40, 40, 255} |

## API
### Base URL
Public (Vercel Serverless Functions)
```
https://api.led.o38.me/
```

Local
```
http://localhost:8080/
```

### Endpoints
- `POST /api/` - Image generation endpoint (accepts image file and settings via form-data)

## License
MIT License - see the [LICENSE](LICENSE) file for details.