# TSGCTF2023 Health Checker

![Lint](https://github.com/tsg-ut/tsgctf-checker/actions/workflows/lint.yml/badge.svg)
![Test](https://github.com/tsg-ut/tsgctf-checker/actions/workflows/test.yml/badge.svg)

## 💻 Usage

### Setup MySQL

```bash
sudo apt install mysql-server
sudo systemctl start mysql.service
sudo mysql -e "source ./scripts/mysql/init.sql"
```

### Create Configuration File

Create configuration JSON file which defines the following variables:

| Key | Type | Description |
|---|---|---|
| `parallel` | int | The number of concurrent test process. |
| `challs_dir` | string | The path to the directory where challenges are placed. |
| `have_genre_dir` | bool | If `false`, directories under `challs_dir` are treated as challenge dir. If `true`, directores under `challs_dir` are treated as genre dir and their sub directories are treated as challenge dir. |
| `targets_file` | string | The path to the file which lists host/port of challenges. |
| `skip_non_exist` | string | Skip challenges who don't have `info.json`. |
| `slack_token` | string (optional) | Slack Bot User OAuth Token. |
| `slack_channel` | string (optional) | Slack channel ID including `#`. |

You can check [the example configuration file](./tests/assets/config.json).

### Create Targets File

Create CSV file which lists host/port of challenges.

- Each row of the file must have three fields:
  - Challenge ID (same as `info.json`'s `name` field)
  - Host name
  - Port
- The file must NOT have header row.

Specify the path to this file to `targets_file` field of the configuration file.

### Setup Environment Variables

| ENV | Description |
|---|---|
| `DBUSER` | Username of MySQL. (checker/badge) |
| `DBPASS` | Password of user `DBUSER`. (checker/badge) |
| `DBHOST` | Host name of MySQL. (checker/badge) |
| `DBNAME` | Database name of MySQL. (checker/badge) |
| `BADGE_PORT` | Port number of badge server. Default to `8080`. (badge) |

### Run and records tests

```bash
make cmd
./bin/cmd/checker --config=<config path>
```

### Run badge server

Badge is served as `/badge/<challenge name>`.

```bash
make cmd
./bin/cmd/badge
# or
./bin/cmd/badge --port=<port number>
```

## 📢 Slack Notification

If your run `checker` with `--notify-slack` option,
failed tests would be notified to Slack.

## 🇯🇵 Challenge Requirement

A directory specified by `challs_dir` looks like the following:

```bash
# When `have_genre_dir` is `false`
`challs_dir`
├── chall1
│   └── solver
│       ├── Dockerfile
│       └── info.json
└── chall2
    └── solver
        ├── Dockerfile
        └── info.json

# When `have_genre_dir` is `true`
`challs_dir`
├── pwn
│   ├── chall1
│   │   └── solver
│   │       ├── Dockerfile
│   │       └── info.json
│   └── chall2
│           ├── Dockerfile
│           └── info.json
└── web
    └── chall3
            ├── Dockerfile
            └── info.json
```

- Each challenge must have `info.json` file.
- Each challenge must have `Dockerfile`.
- Host and port of each challenge are passed to the solver container as 1st/2nd argument.

`info.json` must have the following keys:

| Key | Type | Description |
|---|---|---|
| `name` | string | Unique name of the challenge. Numbers, alphabets, `-`, `_`, and space are allowed. |
| `timeout` | int | Timeout in seconds including the time to build a testing container. |
| `assignee` | string | Slack User ID of the challenge author. Mentioned to on test failure. |

## 😈 Daemonization

TBD

## 🌳 Development

```bash
# Build binaries under `/bin/cmd` directory
make cmd

# Run tests
make test

# Run specific test
go test -v <package dir> -run <test name>

# Format
make fmt
```

### Notes

- If you want to run example challenge and solver on the same local machine, you can pass `--network="host"` option to `/bin/cmd/checker` to make container use host network.
