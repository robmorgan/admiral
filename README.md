# Admiral

A simple CLI tool for working with AWS powered by the AWS Go SDK.

## Features

 - List all running EC2 instances
 - List all running ECS tasks for a specific cluster

## Install

```
$ go get -u -v github.com/robmorgan/admiral
```

## Usage

```
$ admiral hosts list
$ admiral containers list production-app
```

Or use the short-hand syntax:

```
$ admiral h l
$ admiral c l production-app
```

## Contributing

Pull requests are welcome.
