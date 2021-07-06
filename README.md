# Vulcan

## 1. Introduction

Github Actions is awesome but it is not completely free and it requires Github environment. Vulcan is created to solve these issue. It is inspired by Github Actions, so you may figure out few similarity between their configuration syntax, but they may are different in mechanism and concept.

Vulcan is a container-based build tool which run build process inside a docker container. It brings to you the independent and consistent build environment.

With Vulcan, you may:
- Be able to customize build tool
- Take less cost due to it is open source and free
- Build your project in local and other SCM systems outside Github

> Why Vulcan?
>
> Vulcan is a god of fire in Roman mythology. His Greek equivalent is the god Hephaestus. He is a very talented blacksmith. Through his identification with the Hephaestus of Greek mythology, he came to be considered as the manufacturer of art, arms, iron, jewelry and armor for various gods and heroes, including the thunderbolts of Jupiter (Zeus).

## 2. Roadmap

- Add step building controller (Vulcan Executor)
- Add capability to support external plugin
- Add capability to be deployed as a build system

## 3. How to install

```bash
$ git clone git@github.com:locngoxuan/vulcan.git

$ cd vulcan

$ make && make install

$ export VULCAN_HOME=~/.vulcan

$ export PATH=$VULCAN_HOME/bin:$PATH

$ source .

$ vlocal --action example [--job set-var] //empty is all jobs
```