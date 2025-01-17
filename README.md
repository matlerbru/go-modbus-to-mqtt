# Modbus-to-MQTT

This project is a Go application that reads data from Modbus devices and publishes it to an MQTT broker.
The publishing format is templateable using go template library.

## Features

- Read data from Modbus devices
- Publish data to an MQTT broker
- Configurable via a YAML file

## Requirements

- Go 1.16 or later
- Modbus device
- MQTT broker

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/go-modbus-to-mqtt.git
    ```
2. Navigate to the project directory:
    ```sh
    cd go-modbus-to-mqtt
    ```
3. Build the application:
    ```sh
    go build
    ```

## Configuration

Create a `config.yaml` file in the project directory, 
or set the config location using the `MODBUS_TO_MQTT_CONFIG_PATH` environment variable

The configuraion consists of:

| Section   | Parameter            | Description                                                                                          | Type      | Example Value                   |
|-----------|----------------------|------------------------------------------------------------------------------------------------------|-----------|---------------------------------|
| **mqtt**  | `address`            | The IP address of the MQTT broker.                                                                   | string    | "192.168.1.206"                 |
|           | `port`               | The port number for the MQTT broker connection.                                                      | integer   | 1883                            |
|           | `qos`                | The Quality of Service level for the MQTT messages.                                                  | integer   | 2                               |
|           | `mainTopic`          | The main topic to be used for communication with the MQTT broker.                                    | string    | "modbus-to-mqtt"                |
| **modbus**| `address`            | The IP address of the Modbus device.                                                                 | string    | "192.168.1.12"                  |
|           | `port`               | The port number for the Modbus connection.                                                           | integer   | 502                             |
|           | `blocks`             | A list of Modbus blocks to define, each specifying parameters for Modbus communication.              | list      | See below                       |
|           | `scanInterval`       | The interval (in milliseconds) at which to scan the Modbus device for new data.                      | integer   | 80                              |
| **metrics**| `enabled`           | Whether metrics collection is enabled.                                                               | boolean   | true                            |

#### Modbus Block Configuration

| Parameter            | Description                                                                                          | Type      | Example Value                   |
|----------------------|------------------------------------------------------------------------------------------------------|-----------|---------------------------------|
| **type**             | The type of Modbus block (e.g., `coil`).                                                             | string    | "coil"                          |
| **start**            | The starting address for the Modbus block.                                                           | integer   | 0                               |
| **count**            | The number of items to read or write in the Modbus block.                                            | integer   | 1                               |
| **topic**            | The topic for the MQTT message, can include variables (e.g., `{{.Address}}`).                        | string    | "di/{{.Address}}"               |
| **report**           | A list of conditions under which data should be sent, including format and state change logic.       | list      | See below                       |

#### Report Configuration

| Parameter            | Description                                                                                          | Type      | Example Value                   |
|----------------------|------------------------------------------------------------------------------------------------------|-----------|-----------------------------------|
| **sendOn**           | A condition in Go templating syntax that determines when to send data.                               | string    | `{{(gt .State.LastChanged 999)}}` |
| **format**           | The format of the message (e.g., `long`, `short`).                                                   | string    | "long"                            |
| **onlyOnChange**     | Whether to send the report only if the state has changed.                                            | boolean   | true                              |

## Templating

The Go structure used for templating is as follows:

| Variable            | Description                                                                                          | Example Value           |
|---------------------|------------------------------------------------------------------------------------------------------|-------------------------|
| `State`             | Holds the current state of the Modbus block.                                                         | See `State` struct below |
| `Address`           | The address of the Modbus device/block.                                                              | `0`                     |
| `ScanInterval`      | The scan interval (in milliseconds) for reading data from the Modbus device.                         | `80`                    |
| `BlockStartAddress` | The starting address for the Modbus block.                                                           | `0`                     |
| `BlockAddressCount` | The number of addresses in the Modbus block.                                                         | `1`                     |
| `BlockType`         | The type of Modbus block (e.g., `coil`).                                                             | `"coil"`                |

#### State struct

| Variable      | Description                                                              | Example Value       |
|---------------|--------------------------------------------------------------------------|---------------------|
| `Value`       | The current value of the Modbus address.                                 | `true` / `42`       |
| `LastChanged` | The timestamp when the state was last updated.                           | `1000`              |
| `Changed`     | Indicates whether the state has changed since the last check.            | `true`              |

## Usage

Run the application:
```sh
./go-modbus-to-mqtt
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Disclaimer

This is a small project created for learning Go and experimenting with Modbus and MQTT integration. 
The project is a work in progress, and currently, only input coils are implemented. 
However, any suggestions or contributions to improve the project are very welcome!

## Acknowledgements

- [Go Modbus Library](https://github.com/goburrow/modbus)
- [Paho MQTT Go Client](https://github.com/eclipse/paho.mqtt.golang)
