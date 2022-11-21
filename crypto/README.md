# Build a Command Line Interface APP with Golang

Requirements:
- Go
- Docker

If it's not done yet , you can download and install Go here [golang.org](https://go.dev)

For this project , we use [urfave/cli](https://pkg.go.dev/github.com/urfave/cli/v2) package

Our cli app will be very simple , all we want to display is:
- The name of our app
- A short description
- List of flags
- List of commands
- help Usage

All these info are in the [app.go](https://github.com/yvesDenis/website-projects-articles/blob/crypto/crypto/cli-app/app.go) file.

### Commands

Commands are the actaions performed by the cli. For the use case, we want our app to encrypt and decrypt data with a key provided by the user.
The encryption mechanism is implemented in the [crypto-example](https://github.com/yvesDenis/website-projects-articles/blob/crypto/crypto/crypto-example/encryption.go) file.
We are 2 commands for our CLI, one for the encryption:

```json
    {
        Name:    "encrypt",
        Aliases: []string{"e"},
        Usage:   "encrypt a text",
        Action: func(*cli.Context) error {
            cypherText, err := encryption.Encrypt(text, key)
            if err != nil {
                return cli.Exit("Input parameters are not valid for encryption, please provide valid key or text!", 86)
            }
            fmt.Println(cypherText)
            return nil
        },
    }
```

The other for the decryption :

```json
    {
        Name:    "decrypt",
        Aliases: []string{"d"},
        Usage:   "decrypt a text",
        Action: func(*cli.Context) error {
            plainText, err := encryption.Decrypt(text, key)
            if err != nil {
                return cli.Exit("Input parameters are not valid for decryption, please provide valid key or text!", 86)
            }
            fmt.Println(plainText)
            return nil
        },
    }
```

### Flags

Flags are external variable data passed alongside the commands necessary to perform actions.
The app takes 2 required values: **The cypherKey** and **the text to encrypt/decrypt**:


```
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:        "key",
            Aliases:     []string{"k"},
            Usage:       "(Required)-Cypher key used for encryption/decryption, It's length should be a multiple of 16",
            Destination: &key,
        },
        &cli.StringFlag{
            Name:        "text",
            Aliases:     []string{"t"},
            Usage:       "(Required)-Text to encrypt/decrypt",
            Destination: &text,
        },
    },

```

### Usage
- If you run locally , you need to build/install the program and afterwards run it:

```
    go build 

    app -h

```

- A docker image of this app is already built and available on dockerHub: https://hub.docker.com/repository/docker/yvesdeffo/app-cli

Docker command for encryption : docker run yvesdeffo/app-cli -k "Your-key" -t"Your-Text" encrypt

Docker command for decryption : docker run yvesdeffo/app-cli -k "Your-key" -t"Your-Text" decrypt








